; indentation queries for mon_lang (Neovim nvim-treesitter)
; SPDX-License-Identifier: MIT

[
  (block)
  (parameter_list)
  (argument_list)
  (struct_declaration)
  (impl_block)
  (match_statement)
] @indent.begin

["}" ")" "]"] @indent.branch
"}" @indent.end
