# tree-sitter-mon

A [tree-sitter](https://tree-sitter.github.io/) grammar for **mon_lang**, the
Mongolian-keyword systems language in this repository. It powers syntax
highlighting, structural editing, and the language server across editors
(Neovim, Helix, and any tree-sitter host).

## Layout

```
grammar.js            the grammar
tree-sitter.json      grammar metadata (scope, file types, query paths)
queries/
  highlights.scm      syntax highlighting
  locals.scm          scopes & definitions (rename, go-to-def)
  indents.scm         indentation (Neovim)
test/corpus/          parser regression tests
src/                  generated parser (parser.c, node-types.json)
```

## Building

```sh
tree-sitter generate      # regenerate src/ from grammar.js
tree-sitter test          # run the corpus tests
tree-sitter parse FILE.mn # dump a parse tree
```

## Coverage

Functions (and forward prototypes), `extern`, `тунх` (public), structs and
`хэрэгжүүл` impl blocks with `өөрөө` receivers, `зарла` declarations, arrays
(`тоо64[]`, sized `тоо64[24]`), pointer types (`тоо*`) with `&`/`*`, `шинэ`
allocation, casts (`гэж`), `хэрэв/эсвэл`, `давтах` while-loops, `давт … хүртэл`
range loops, `тааруул` match with wildcard `_`, `буц/зогс/үргэлжлүүл`, module
imports (`ашигла "…" [гэж alias]`), string/char/number/float literals with
escapes, and line comments. Cyrillic identifiers are first-class.

The grammar is validated against the whole in-repo corpus: every program under
`tests/run/`, the self-hosted compiler (`selfhost/*.mn`), and the standard
library (`stdlib/**`) parses without error.
