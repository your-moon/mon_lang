-- Neovim integration for mon_lang: tree-sitter parser registration and,
-- optionally, the language server.
--
-- Usage (lazy.nvim):
--   { dir = "/path/to/mon_lang/tools/nvim-mon",
--     dependencies = { "nvim-treesitter/nvim-treesitter" },
--     config = function() require("mon").setup() end }

local M = {}

-- Absolute path to the tree-sitter-mon grammar shipped alongside this plugin
-- (../../tree-sitter-mon relative to this file).
local function grammar_dir()
  local here = debug.getinfo(1, "S").source:sub(2)          -- .../lua/mon/init.lua
  local plugin = vim.fn.fnamemodify(here, ":h:h:h")          -- .../nvim-mon
  return vim.fn.fnamemodify(plugin .. "/../tree-sitter-mon", ":p")
end

--- Register the mon parser with nvim-treesitter and wire up highlighting.
--- @param opts table|nil  { grammar = <path>, lsp = <bool>, cmd = <string> }
function M.setup(opts)
  opts = opts or {}

  local ok, parsers = pcall(require, "nvim-treesitter.parsers")
  if ok then
    local configs = parsers.get_parser_configs and parsers.get_parser_configs() or parsers
    configs.mon = {
      install_info = {
        url = opts.grammar or grammar_dir(),
        files = { "src/parser.c" },
        generate_requires_npm = false,
        requires_generate_from_grammar = false,
      },
      filetype = "mon",
    }
  end

  -- Map filetype -> parser language so :TSBufEnable highlight works.
  pcall(function()
    vim.treesitter.language.register("mon", "mon")
  end)

  if opts.lsp then
    M.setup_lsp(opts)
  end
end

--- Start the mon_lang language server for mon buffers.
--- @param opts table|nil  { cmd = <string|table> }
function M.setup_lsp(opts)
  opts = opts or {}
  local cmd = opts.cmd or { "mon-lsp" }
  if type(cmd) == "string" then
    cmd = { cmd }
  end

  vim.api.nvim_create_autocmd("FileType", {
    pattern = "mon",
    callback = function(args)
      vim.lsp.start({
        name = "mon-lsp",
        cmd = cmd,
        root_dir = vim.fs.root(args.buf, { ".git", "go.mod" })
          or vim.fn.getcwd(),
      })
    end,
  })
end

return M
