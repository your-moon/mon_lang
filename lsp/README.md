# mon-lsp

A Language Server for **mon_lang**, built directly on the compiler's own
lexer, parser, and semantic analyser — no third-party LSP library, matching the
project's dependency-free philosophy.

## Build

```sh
go build -o mon-lsp ./cmd/mon-lsp
```

Put the resulting `mon-lsp` on your `PATH`.

## Features

| Capability                    | Source of truth                                  |
| ----------------------------- | ------------------------------------------------ |
| Diagnostics (errors)          | the real lex → parse → semantic pipeline         |
| Document symbols              | functions, structs (+fields), globals, methods   |
| Hover                         | one-line signature of the symbol under the cursor|
| Go-to-definition              | jumps to a top-level declaration by name         |

Diagnostics run the same front end the compiler uses, so anything that would
fail `mon gen` is reported inline, positioned by the line the compiler names.

## Editor setup

- **Neovim**: `require("mon").setup({ lsp = true })` — see `tools/nvim-mon`.
- **Generic LSP client**: launch `mon-lsp` over stdio for `filetype = mon`.

## Protocol

Speaks LSP over stdio with `Content-Length` framing (`lsp/protocol.go`), full
document sync. The message loop and handlers live in `lsp/server.go`; the
`cmd/mon-lsp` binary is a thin `stdin`/`stdout` wrapper.
