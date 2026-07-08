/*
 * mon_lang - lsp
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package lsp

import (
	"encoding/json"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/your-moon/mon_lang/lexer"
	"github.com/your-moon/mon_lang/mtypes"
	"github.com/your-moon/mon_lang/parser"
	semanticanalysis "github.com/your-moon/mon_lang/semantic_analysis"
	"github.com/your-moon/mon_lang/symbols"
	"github.com/your-moon/mon_lang/util/unique"
)

// document is one open source buffer plus its derived index.
type document struct {
	uri   string
	text  string
	runes []int32
	lines []int // rune offset of the start of each line
	prog  *parser.ASTProgram
}

// Server is a single-client LSP server over stdio.
type Server struct {
	conn *conn
	docs map[string]*document
}

// NewServer wires the server to a byte stream (usually stdin/stdout).
func NewServer(in io.Reader, out io.Writer) *Server {
	return &Server{
		conn: newConn(in, out),
		docs: make(map[string]*document),
	}
}

// Run pumps the message loop until the stream closes or `exit` arrives.
func (s *Server) Run() error {
	for {
		msg, err := s.conn.read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if s.dispatch(msg) {
			return nil // clean exit requested
		}
	}
}

// dispatch handles one message; returns true when the server should stop.
func (s *Server) dispatch(msg *rpcMessage) bool {
	switch msg.Method {
	case "initialize":
		s.handleInitialize(msg)
	case "initialized":
		// notification, nothing to do
	case "textDocument/didOpen":
		s.handleDidOpen(msg)
	case "textDocument/didChange":
		s.handleDidChange(msg)
	case "textDocument/didClose":
		s.handleDidClose(msg)
	case "textDocument/documentSymbol":
		s.handleDocumentSymbol(msg)
	case "textDocument/hover":
		s.handleHover(msg)
	case "textDocument/definition":
		s.handleDefinition(msg)
	case "shutdown":
		s.conn.respond(msg.ID, nil)
	case "exit":
		return true
	default:
		// Unknown request: reply with an empty result so the client isn't
		// left waiting. Unknown notifications (no id) are ignored.
		if len(msg.ID) > 0 {
			s.conn.respond(msg.ID, nil)
		}
	}
	return false
}

func (s *Server) handleInitialize(msg *rpcMessage) {
	capabilities := map[string]interface{}{
		"textDocumentSync":       1, // full document sync
		"documentSymbolProvider": true,
		"hoverProvider":          true,
		"definitionProvider":     true,
	}
	result := map[string]interface{}{
		"capabilities": capabilities,
		"serverInfo":   map[string]string{"name": "mon-lsp", "version": "0.1.0"},
	}
	s.conn.respond(msg.ID, result)
}

// ---- document lifecycle ----

func (s *Server) handleDidOpen(msg *rpcMessage) {
	var p didOpenParams
	if err := json.Unmarshal(msg.Params, &p); err != nil {
		return
	}
	s.updateDoc(p.TextDocument.URI, p.TextDocument.Text)
}

func (s *Server) handleDidChange(msg *rpcMessage) {
	var p didChangeParams
	if err := json.Unmarshal(msg.Params, &p); err != nil {
		return
	}
	if len(p.ContentChanges) == 0 {
		return
	}
	// full sync: the last change carries the whole document
	s.updateDoc(p.TextDocument.URI, p.ContentChanges[len(p.ContentChanges)-1].Text)
}

func (s *Server) handleDidClose(msg *rpcMessage) {
	var p didCloseParams
	if err := json.Unmarshal(msg.Params, &p); err != nil {
		return
	}
	delete(s.docs, p.TextDocument.URI)
}

// updateDoc re-indexes a buffer and republishes diagnostics.
func (s *Server) updateDoc(uri, text string) {
	runes := runesOf(text)
	doc := &document{
		uri:   uri,
		text:  text,
		runes: runes,
		lines: lineStarts(runes),
	}
	// Parse once for the symbol index; parse errors don't stop indexing.
	prog, _ := parser.NewParser(runes).ParseProgram()
	doc.prog = prog
	s.docs[uri] = doc

	s.publishDiagnostics(doc)
}

// ---- diagnostics: run the real front end and surface the first failure ----

var lineInMessage = regexp.MustCompile(`(\d+)-р мөр`)

func (s *Server) publishDiagnostics(doc *document) {
	diags := s.analyze(doc)
	s.conn.notify("textDocument/publishDiagnostics", publishDiagnosticsParams{
		URI:         doc.uri,
		Diagnostics: diags,
	})
}

func (s *Server) analyze(doc *document) []Diagnostic {
	// 1) lexer — precise line/column
	scanner := lexer.NewScanner(doc.runes)
	for {
		tok, err := scanner.Scan()
		if err != nil {
			return []Diagnostic{s.messageDiagnostic(doc, err.Error())}
		}
		if tok.Type == lexer.EOF {
			break
		}
	}

	// 2) parser
	p := parser.NewParser(doc.runes)
	node, err := p.ParseProgram()
	if err != nil {
		return []Diagnostic{s.messageDiagnostic(doc, err.Error())}
	}
	if errs := p.Errors(); len(errs) > 0 {
		out := make([]Diagnostic, 0, len(errs))
		for _, e := range errs {
			out = append(out, s.messageDiagnostic(doc, e.Error()))
		}
		return out
	}

	// 3) semantic analysis — reuse the embedded prelude/std packages
	baseDir := "."
	if path := uriToPath(doc.uri); path != "" {
		baseDir = filepath.Dir(path)
	}
	table := symbols.NewSymbolTable()
	analyzer := semanticanalysis.NewSemanticAnalyzer(doc.runes, unique.NewUniqueGen(), table, baseDir, "")
	if _, _, err := analyzer.Analyze(node); err != nil {
		return []Diagnostic{s.messageDiagnostic(doc, err.Error())}
	}

	return []Diagnostic{}
}

// messageDiagnostic turns a compiler error string into a positioned
// diagnostic, extracting a line number from the message when present.
func (s *Server) messageDiagnostic(doc *document, message string) Diagnostic {
	line := 0
	if m := lineInMessage.FindStringSubmatch(message); m != nil {
		// compiler lines are 1-based; LSP lines are 0-based
		if n := atoi(m[1]); n > 0 {
			line = n - 1
		}
	}
	rng := Range{
		Start: Position{Line: line, Character: 0},
		End:   Position{Line: line, Character: s.lineLength(doc, line)},
	}
	return Diagnostic{
		Range:    rng,
		Severity: severityError,
		Source:   "mon",
		Message:  strings.TrimSpace(message),
	}
}

func (s *Server) lineLength(doc *document, line int) int {
	if line < 0 || line >= len(doc.lines) {
		return 0
	}
	start := doc.lines[line]
	end := len(doc.runes)
	if line+1 < len(doc.lines) {
		end = doc.lines[line+1]
	}
	n := end - start
	if n > 0 {
		return n - 1 // drop the trailing newline
	}
	return 0
}

// ---- document symbols ----

func (s *Server) handleDocumentSymbol(msg *rpcMessage) {
	var p documentSymbolParams
	if err := json.Unmarshal(msg.Params, &p); err != nil {
		s.conn.respond(msg.ID, []DocumentSymbol{})
		return
	}
	doc := s.docs[p.TextDocument.URI]
	if doc == nil || doc.prog == nil {
		s.conn.respond(msg.ID, []DocumentSymbol{})
		return
	}
	s.conn.respond(msg.ID, s.symbols(doc))
}

func (s *Server) symbols(doc *document) []DocumentSymbol {
	var out []DocumentSymbol
	for _, decl := range doc.prog.Decls {
		switch d := decl.(type) {
		case *parser.FnDecl:
			kind := symKindFunction
			if d.IsMethod {
				kind = symKindMethod
			}
			out = append(out, DocumentSymbol{
				Name:           d.Ident,
				Detail:         fnSignature(d),
				Kind:           kind,
				Range:          s.tokenRange(doc, d.Token),
				SelectionRange: s.tokenRange(doc, d.Token),
			})
		case *parser.VarDecl:
			out = append(out, DocumentSymbol{
				Name:           d.Ident,
				Detail:         typeString(d.VarType),
				Kind:           symKindVariable,
				Range:          s.tokenRange(doc, d.Token),
				SelectionRange: s.tokenRange(doc, d.Token),
			})
		case *parser.ASTStructDecl:
			var fields []DocumentSymbol
			for i, fname := range d.FieldNames {
				detail := ""
				if i < len(d.FieldTypes) {
					detail = typeString(d.FieldTypes[i])
				}
				fields = append(fields, DocumentSymbol{
					Name:           fname,
					Detail:         detail,
					Kind:           symKindField,
					Range:          s.tokenRange(doc, d.Token),
					SelectionRange: s.tokenRange(doc, d.Token),
				})
			}
			out = append(out, DocumentSymbol{
				Name:           d.Name,
				Kind:           symKindStruct,
				Range:          s.tokenRange(doc, d.Token),
				SelectionRange: s.tokenRange(doc, d.Token),
				Children:       fields,
			})
		}
	}
	return out
}

// ---- hover & go-to-definition, keyed by the identifier under the cursor ----

func (s *Server) handleHover(msg *rpcMessage) {
	var p textDocumentPositionParams
	if err := json.Unmarshal(msg.Params, &p); err != nil {
		s.conn.respond(msg.ID, nil)
		return
	}
	doc := s.docs[p.TextDocument.URI]
	if doc == nil || doc.prog == nil {
		s.conn.respond(msg.ID, nil)
		return
	}
	word := s.wordAt(doc, p.Position)
	if word == "" {
		s.conn.respond(msg.ID, nil)
		return
	}
	if sig := s.signatureOf(doc, word); sig != "" {
		s.conn.respond(msg.ID, Hover{
			Contents: MarkupContent{Kind: "markdown", Value: "```mon\n" + sig + "\n```"},
		})
		return
	}
	s.conn.respond(msg.ID, nil)
}

func (s *Server) handleDefinition(msg *rpcMessage) {
	var p textDocumentPositionParams
	if err := json.Unmarshal(msg.Params, &p); err != nil {
		s.conn.respond(msg.ID, nil)
		return
	}
	doc := s.docs[p.TextDocument.URI]
	if doc == nil || doc.prog == nil {
		s.conn.respond(msg.ID, nil)
		return
	}
	word := s.wordAt(doc, p.Position)
	if word == "" {
		s.conn.respond(msg.ID, nil)
		return
	}
	if tok, ok := s.declTokenOf(doc, word); ok {
		s.conn.respond(msg.ID, Location{URI: doc.uri, Range: s.tokenRange(doc, tok)})
		return
	}
	s.conn.respond(msg.ID, nil)
}

// signatureOf renders a one-line signature for a top-level name.
func (s *Server) signatureOf(doc *document, name string) string {
	for _, decl := range doc.prog.Decls {
		switch d := decl.(type) {
		case *parser.FnDecl:
			if d.Ident == name {
				return fnSignature(d)
			}
		case *parser.VarDecl:
			if d.Ident == name {
				return "зарла " + d.Ident + ": " + typeString(d.VarType)
			}
		case *parser.ASTStructDecl:
			if d.Name == name {
				return "бүтэц " + d.Name
			}
		}
	}
	return ""
}

// declTokenOf finds the defining token for a top-level name.
func (s *Server) declTokenOf(doc *document, name string) (lexer.Token, bool) {
	for _, decl := range doc.prog.Decls {
		switch d := decl.(type) {
		case *parser.FnDecl:
			if d.Ident == name {
				return d.Token, true
			}
		case *parser.VarDecl:
			if d.Ident == name {
				return d.Token, true
			}
		case *parser.ASTStructDecl:
			if d.Name == name {
				return d.Token, true
			}
		}
	}
	return lexer.Token{}, false
}

// ---- position helpers ----

func (s *Server) tokenRange(doc *document, tok lexer.Token) Range {
	start := s.offsetToPos(doc, tok.Span.Start)
	end := s.offsetToPos(doc, tok.Span.End)
	if tok.Span.End <= tok.Span.Start {
		end = start
	}
	return Range{Start: start, End: end}
}

// offsetToPos maps a rune offset to an LSP position. Cyrillic lives in the
// BMP, so one rune equals one UTF-16 code unit for mon_lang source.
func (s *Server) offsetToPos(doc *document, offset int) Position {
	if offset < 0 {
		offset = 0
	}
	// binary-search-free linear scan: line count is small for source files
	line := 0
	for i := 1; i < len(doc.lines); i++ {
		if doc.lines[i] > offset {
			break
		}
		line = i
	}
	return Position{Line: line, Character: offset - doc.lines[line]}
}

// posToOffset is the inverse: an LSP position to a rune offset.
func (s *Server) posToOffset(doc *document, pos Position) int {
	if pos.Line < 0 || pos.Line >= len(doc.lines) {
		return 0
	}
	return doc.lines[pos.Line] + pos.Character
}

// wordAt returns the identifier spanning the cursor, or "".
func (s *Server) wordAt(doc *document, pos Position) string {
	off := s.posToOffset(doc, pos)
	if off < 0 || off > len(doc.runes) {
		return ""
	}
	lo := off
	for lo > 0 && isIdentRune(doc.runes[lo-1]) {
		lo--
	}
	hi := off
	for hi < len(doc.runes) && isIdentRune(doc.runes[hi]) {
		hi++
	}
	if lo == hi {
		return ""
	}
	return string(doc.runes[lo:hi])
}

// ---- small utilities ----

// isIdentRune mirrors the lexer: ASCII letters/digits, '_', and any
// non-ASCII rune (which covers the Cyrillic identifiers mon_lang uses).
func isIdentRune(r int32) bool {
	if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
		return true
	}
	if r >= '0' && r <= '9' || r == '_' {
		return true
	}
	return r >= 128
}

func runesOf(text string) []int32 {
	out := make([]int32, 0, len(text))
	for len(text) > 0 {
		r, size := utf8.DecodeRuneInString(text)
		out = append(out, r)
		text = text[size:]
	}
	return out
}

func lineStarts(runes []int32) []int {
	lines := []int{0}
	for i, r := range runes {
		if r == '\n' {
			lines = append(lines, i+1)
		}
	}
	return lines
}

func fnSignature(d *parser.FnDecl) string {
	var b strings.Builder
	if d.IsExtern {
		b.WriteString("extern ")
	}
	if d.IsPublic {
		b.WriteString("тунх ")
	}
	b.WriteString("функц ")
	b.WriteString(d.Ident)
	b.WriteByte('(')
	for i, param := range d.Params {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(param.Ident)
		b.WriteByte(' ')
		b.WriteString(typeString(param.Type))
	}
	b.WriteString(") -> ")
	b.WriteString(typeString(d.ReturnType))
	return b.String()
}

// typeString renders a mon_lang type using its source keywords.
func typeString(t mtypes.Type) string {
	switch ty := t.(type) {
	case nil:
		return "?"
	case *mtypes.VoidType:
		return "хоосон"
	case *mtypes.Int32Type:
		return "тоо"
	case *mtypes.Int64Type:
		return "тоо64"
	case *mtypes.UInt32Type:
		return "этоо"
	case *mtypes.UInt64Type:
		return "этоо64"
	case *mtypes.Float64Type:
		return "бутархай"
	case *mtypes.StringType:
		return "мөр"
	case *mtypes.ArrayType:
		return typeString(ty.ElementType) + "[]"
	case *mtypes.PointerType:
		return typeString(ty.Referenced) + "*"
	case *mtypes.NamedType:
		return ty.Name
	case *mtypes.StructType:
		return ty.Name
	case *mtypes.FnType:
		return "функц"
	default:
		return "?"
	}
}

func uriToPath(uri string) string {
	const prefix = "file://"
	if strings.HasPrefix(uri, prefix) {
		return uri[len(prefix):]
	}
	return ""
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return n
		}
		n = n*10 + int(c-'0')
	}
	return n
}
