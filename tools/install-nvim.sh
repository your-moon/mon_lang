#!/bin/sh
# Install / refresh mon_lang support in the user's Neovim config:
#   - build the tree-sitter parser and drop it on the runtimepath
#   - install the highlight/locals/indent queries
#   - build the mon-lsp language server
#   - install ftplugin/mon.lua (tree-sitter highlighting + LSP)
#
# Re-run this after changing the grammar (tools/tree-sitter-mon) or the
# language server (cmd/mon-lsp, lsp/).
set -e

repo=$(cd "$(dirname "$0")/.." && pwd)
nvim_cfg="${NVIM_CONFIG:-$HOME/.config/nvim}"

echo "repo:       $repo"
echo "nvim config: $nvim_cfg"

# 1. language server
echo "building mon-lsp..."
(cd "$repo" && go build -o bin/mon-lsp ./cmd/mon-lsp)

# 2. tree-sitter parser
echo "building tree-sitter parser..."
mkdir -p "$nvim_cfg/parser" "$nvim_cfg/queries/mon"
(cd "$repo/tools/tree-sitter-mon" && tree-sitter build -o "$nvim_cfg/parser/mon.so")

# 3. queries
cp "$repo/tools/tree-sitter-mon/queries/highlights.scm" "$nvim_cfg/queries/mon/"
cp "$repo/tools/tree-sitter-mon/queries/locals.scm"     "$nvim_cfg/queries/mon/"
cp "$repo/tools/tree-sitter-mon/queries/indents.scm"    "$nvim_cfg/queries/mon/"

# 4. ftplugin (points the LSP at the freshly built binary)
mkdir -p "$nvim_cfg/ftplugin"
cat > "$nvim_cfg/ftplugin/mon.lua" <<EOF
-- mon_lang: native tree-sitter highlighting + mon-lsp. Managed by
-- $repo/tools/install-nvim.sh — re-run that to refresh.
vim.bo.commentstring = "// %s"
vim.bo.comments = "://"
vim.bo.expandtab = true
vim.bo.shiftwidth = 4
vim.bo.tabstop = 4
vim.bo.softtabstop = 4

pcall(vim.treesitter.start, 0, "mon")

local lsp_bin = "$repo/bin/mon-lsp"
if vim.fn.executable(lsp_bin) == 1 then
  vim.lsp.start({
    name = "mon-lsp",
    cmd = { lsp_bin },
    root_dir = vim.fs.root(0, { ".git", "go.mod" }) or vim.fn.getcwd(),
  })
end
EOF

# 5. filetype detection (in case it isn't set up already)
mkdir -p "$nvim_cfg/ftdetect"
if [ ! -f "$nvim_cfg/ftdetect/mon.vim" ] && [ ! -f "$nvim_cfg/ftdetect/mon.lua" ]; then
  echo 'au BufRead,BufNewFile *.mn set filetype=mon' > "$nvim_cfg/ftdetect/mon.vim"
fi

echo "done. Open a .mn file in Neovim — tree-sitter highlighting and mon-lsp"
echo "will be active. (Diagnostics display depends on your vim.diagnostic.config.)"
