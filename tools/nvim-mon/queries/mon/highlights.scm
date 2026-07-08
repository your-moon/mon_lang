; tree-sitter highlight queries for mon_lang
; SPDX-License-Identifier: MIT

; Generic fallback FIRST: Neovim/Helix use last-match-wins, so the specific
; captures further down override this for names they recognise.
(identifier) @variable

; ---- comments ----
(line_comment) @comment

; ---- literals ----
(number_literal) @number
(float_literal) @number.float
(char_literal) @character
(string_literal) @string
(escape_sequence) @string.escape

; ---- types ----
(primitive_type) @type.builtin
(struct_declaration name: (identifier) @type)
(impl_block type: (identifier) @type)
(field_declaration name: (identifier) @property)
(field_expression field: (identifier) @property)

; a bare identifier used as a type (struct name in an annotation)
(parameter type: (identifier) @type)
(local_var_declaration type: (identifier) @type)
(global_var_declaration type: (identifier) @type)
(field_declaration type: (identifier) @type)
(cast_expression type: (identifier) @type)
(array_type (identifier) @type)
(pointer_type (identifier) @type)
(new_expression type: (identifier) @type)

; ---- functions ----
(function_declaration name: (identifier) @function)
(extern_declaration name: (identifier) @function)
(call_expression function: (identifier) @function.call)
(method_call_expression method: (identifier) @function.method.call)

; ---- parameters ----
(parameter name: (identifier) @variable.parameter)
(self_parameter) @variable.builtin
(self_expression) @variable.builtin

; ---- keywords ----
[
  "функц"
  "зарла"
  "буц"
  "шинэ"
  "гэж"
] @keyword

[
  "хэрэв"
  "бол"
  "эсвэл"
  "тааруул"
  "=>"
] @keyword.conditional

[
  "давтах"
  "давт"
  "хүртэл"
  "зогс"
  "үргэлжлүүл"
] @keyword.repeat

[
  "бүтэц"
  "хэрэгжүүл"
] @keyword.type

[
  "ашигла"
  "тунх"
  "extern"
  "статик"
] @keyword.import

(wildcard_pattern) @constant.builtin

; ---- entry point gets special emphasis ----
((function_declaration name: (identifier) @function.builtin)
 (#eq? @function.builtin "үндсэн"))

; ---- operators ----
[
  "+" "-" "*" "/" "%"
  "==" "!=" "<" ">" "<=" ">="
  "&&" "||" "!" "~" "&"
  "=" ".."
] @operator

"->" @punctuation.delimiter

; ---- punctuation ----
["(" ")" "{" "}" "[" "]"] @punctuation.bracket
["," ";" ":" "."] @punctuation.delimiter
