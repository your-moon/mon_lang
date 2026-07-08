/*
 * mon_lang - lsp
 *
 * A dependency-free Language Server Protocol implementation over stdio,
 * reusing the compiler's own lexer, parser, and semantic analyser.
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ---- JSON-RPC 2.0 envelopes ----

type rpcMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ---- LSP core types ----

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Source   string `json:"source"`
	Message  string `json:"message"`
}

const (
	severityError   = 1
	severityWarning = 2
	severityInfo    = 3
	severityHint    = 4
)

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type TextDocumentItem struct {
	URI     string `json:"uri"`
	Text    string `json:"text"`
	Version int    `json:"version"`
}

type textDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type didOpenParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type contentChange struct {
	Text string `json:"text"`
}

type didChangeParams struct {
	TextDocument struct {
		URI     string `json:"uri"`
		Version int    `json:"version"`
	} `json:"textDocument"`
	ContentChanges []contentChange `json:"contentChanges"`
}

type didCloseParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type documentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DocumentSymbol (hierarchical) — https://microsoft.github.io/language-server-protocol
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           int              `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

const (
	symKindFunction = 12
	symKindVariable = 13
	symKindStruct   = 23
	symKindField    = 8
	symKindMethod   = 6
)

type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type publishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// ---- transport: Content-Length framed JSON-RPC over a stream ----

type conn struct {
	r *bufio.Reader
	w io.Writer
}

func newConn(r io.Reader, w io.Writer) *conn {
	return &conn{r: bufio.NewReader(r), w: w}
}

// read pulls the next framed message; io.EOF signals a closed stream.
func (c *conn) read() (*rpcMessage, error) {
	var length int
	for {
		line, err := c.r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break // end of headers
		}
		if strings.HasPrefix(strings.ToLower(line), "content-length:") {
			v := strings.TrimSpace(line[len("content-length:"):])
			length, err = strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("bad Content-Length: %v", err)
			}
		}
	}
	if length == 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(c.r, buf); err != nil {
		return nil, err
	}
	var msg rpcMessage
	if err := json.Unmarshal(buf, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (c *conn) write(msg *rpcMessage) error {
	msg.JSONRPC = "2.0"
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(c.w, "Content-Length: %d\r\n\r\n%s", len(body), body)
	return err
}

// respond replies to a request with a result (or an error).
func (c *conn) respond(id json.RawMessage, result interface{}) error {
	return c.write(&rpcMessage{ID: id, Result: result})
}

func (c *conn) respondError(id json.RawMessage, code int, message string) error {
	return c.write(&rpcMessage{ID: id, Error: &rpcError{Code: code, Message: message}})
}

// notify sends a server->client notification (no id).
func (c *conn) notify(method string, params interface{}) error {
	raw, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return c.write(&rpcMessage{Method: method, Params: raw})
}
