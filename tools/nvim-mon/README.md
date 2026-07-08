# nvim-mon

Neovim support for **mon_lang**: filetype detection, tree-sitter highlighting,
indentation, and an optional language-server hookup.

## Install (lazy.nvim)

```lua
{
  dir = "/path/to/mon_lang/tools/nvim-mon",
  dependencies = { "nvim-treesitter/nvim-treesitter" },
  config = function()
    require("mon").setup({
      -- lsp = true,          -- start mon-lsp for .mn buffers
      -- cmd = { "mon-lsp" }, -- language-server command (see tools/mon-lsp)
    })
  end,
}
```

Then build the parser once:

```
:TSInstall mon
```

`setup()` registers the in-repo `tree-sitter-mon` grammar with nvim-treesitter,
so `:TSInstall mon` compiles it from `../tree-sitter-mon/src/parser.c`.

## What you get

- `.mn` files are detected as filetype `mon`.
- Syntax highlighting, `locals` (scopes for rename / go-to-def), and
  indentation, all driven by the shared tree-sitter queries.
- `commentstring` set to `// %s` so `gcc` / `gc` comment motions work.
- Optional LSP client (`require("mon").setup({ lsp = true })`).

## Queries

`queries/mon/*.scm` are copied verbatim from `../tree-sitter-mon/queries/`.
When the grammar's queries change, re-copy them:

```
cp ../tree-sitter-mon/queries/*.scm queries/mon/
```
