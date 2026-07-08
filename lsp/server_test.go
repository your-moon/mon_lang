/*
 * mon_lang - lsp
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package lsp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// frame wraps a JSON body in an LSP Content-Length envelope.
func frame(body string) string {
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)
}

// drive feeds a sequence of framed messages through a fresh server and
// returns everything the server wrote back, split into JSON bodies.
func drive(t *testing.T, messages ...string) []map[string]any {
	t.Helper()
	var in bytes.Buffer
	for _, m := range messages {
		in.WriteString(frame(m))
	}
	var out bytes.Buffer
	if err := NewServer(&in, &out).Run(); err != nil {
		t.Fatalf("server run: %v", err)
	}
	return parseFrames(t, out.String())
}

func parseFrames(t *testing.T, raw string) []map[string]any {
	t.Helper()
	var msgs []map[string]any
	for len(raw) > 0 {
		i := strings.Index(raw, "\r\n\r\n")
		if i < 0 {
			break
		}
		header := raw[:i]
		raw = raw[i+4:]
		var length int
		for _, line := range strings.Split(header, "\r\n") {
			if strings.HasPrefix(strings.ToLower(line), "content-length:") {
				fmt.Sscanf(strings.TrimSpace(line[len("content-length:"):]), "%d", &length)
			}
		}
		if length > len(raw) {
			break
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(raw[:length]), &m); err != nil {
			t.Fatalf("bad frame: %v", err)
		}
		msgs = append(msgs, m)
		raw = raw[length:]
	}
	return msgs
}

const validProgram = `функц нэмэх(а тоо, б тоо) -> тоо {
    буц а + б;
}

функц үндсэн() -> тоо {
    буц нэмэх(2, 3);
}`

func didOpen(uri, text string) string {
	b, _ := json.Marshal(map[string]any{
		"textDocument": map[string]any{"uri": uri, "text": text, "version": 1},
	})
	return `{"jsonrpc":"2.0","method":"textDocument/didOpen","params":` + string(b) + `}`
}

func request(id int, method string, params any) string {
	b, _ := json.Marshal(params)
	return fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":%q,"params":%s}`, id, method, b)
}

func TestInitializeAdvertisesCapabilities(t *testing.T) {
	msgs := drive(t,
		request(1, "initialize", map[string]any{}),
		`{"jsonrpc":"2.0","method":"exit"}`,
	)
	if len(msgs) == 0 {
		t.Fatal("no response to initialize")
	}
	result := msgs[0]["result"].(map[string]any)
	caps := result["capabilities"].(map[string]any)
	for _, key := range []string{"documentSymbolProvider", "hoverProvider", "definitionProvider"} {
		if caps[key] != true {
			t.Errorf("capability %s not advertised", key)
		}
	}
}

func TestDiagnosticsCleanProgram(t *testing.T) {
	uri := "file:///tmp/ok.mn"
	msgs := drive(t,
		request(1, "initialize", map[string]any{}),
		didOpen(uri, validProgram),
		`{"jsonrpc":"2.0","method":"exit"}`,
	)
	diag := findNotification(msgs, "textDocument/publishDiagnostics")
	if diag == nil {
		t.Fatal("no publishDiagnostics notification")
	}
	params := diag["params"].(map[string]any)
	list := params["diagnostics"].([]any)
	if len(list) != 0 {
		t.Fatalf("clean program reported %d diagnostics: %v", len(list), list)
	}
}

func TestDiagnosticsReportError(t *testing.T) {
	uri := "file:///tmp/bad.mn"
	bad := "функц үндсэн() -> тоо {\n    буц ;\n}\nгарц garbage here"
	msgs := drive(t,
		request(1, "initialize", map[string]any{}),
		didOpen(uri, bad),
		`{"jsonrpc":"2.0","method":"exit"}`,
	)
	diag := findNotification(msgs, "textDocument/publishDiagnostics")
	if diag == nil {
		t.Fatal("no publishDiagnostics notification")
	}
	list := diag["params"].(map[string]any)["diagnostics"].([]any)
	if len(list) == 0 {
		t.Fatal("expected at least one diagnostic for a broken program")
	}
}

func TestDocumentSymbols(t *testing.T) {
	uri := "file:///tmp/sym.mn"
	msgs := drive(t,
		request(1, "initialize", map[string]any{}),
		didOpen(uri, validProgram),
		request(2, "textDocument/documentSymbol", map[string]any{
			"textDocument": map[string]any{"uri": uri},
		}),
		`{"jsonrpc":"2.0","method":"exit"}`,
	)
	resp := findResponse(msgs, 2)
	if resp == nil {
		t.Fatal("no documentSymbol response")
	}
	syms := resp["result"].([]any)
	names := map[string]bool{}
	for _, s := range syms {
		names[s.(map[string]any)["name"].(string)] = true
	}
	if !names["нэмэх"] || !names["үндсэн"] {
		t.Fatalf("expected functions нэмэх and үндсэн, got %v", names)
	}
}

func TestHoverShowsSignature(t *testing.T) {
	uri := "file:///tmp/hov.mn"
	// hover over the call to нэмэх on line 5 (0-based), character 9
	msgs := drive(t,
		request(1, "initialize", map[string]any{}),
		didOpen(uri, validProgram),
		request(3, "textDocument/hover", map[string]any{
			"textDocument": map[string]any{"uri": uri},
			"position":     map[string]any{"line": 5, "character": 9},
		}),
		`{"jsonrpc":"2.0","method":"exit"}`,
	)
	resp := findResponse(msgs, 3)
	if resp == nil || resp["result"] == nil {
		t.Fatal("no hover response")
	}
	contents := resp["result"].(map[string]any)["contents"].(map[string]any)
	value := contents["value"].(string)
	if !strings.Contains(value, "функц нэмэх") {
		t.Fatalf("hover missing signature, got %q", value)
	}
}

func TestGotoDefinition(t *testing.T) {
	uri := "file:///tmp/def.mn"
	msgs := drive(t,
		request(1, "initialize", map[string]any{}),
		didOpen(uri, validProgram),
		request(4, "textDocument/definition", map[string]any{
			"textDocument": map[string]any{"uri": uri},
			"position":     map[string]any{"line": 5, "character": 9},
		}),
		`{"jsonrpc":"2.0","method":"exit"}`,
	)
	resp := findResponse(msgs, 4)
	if resp == nil || resp["result"] == nil {
		t.Fatal("no definition response")
	}
	loc := resp["result"].(map[string]any)
	rng := loc["range"].(map[string]any)
	start := rng["start"].(map[string]any)
	if int(start["line"].(float64)) != 0 {
		t.Fatalf("нэмэх is defined on line 0, definition pointed at line %v", start["line"])
	}
}

func findNotification(msgs []map[string]any, method string) map[string]any {
	for _, m := range msgs {
		if m["method"] == method {
			return m
		}
	}
	return nil
}

func findResponse(msgs []map[string]any, id int) map[string]any {
	for _, m := range msgs {
		if v, ok := m["id"]; ok {
			if f, ok := v.(float64); ok && int(f) == id {
				return m
			}
		}
	}
	return nil
}
