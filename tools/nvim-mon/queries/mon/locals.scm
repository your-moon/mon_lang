; local scope / definition queries for mon_lang
; SPDX-License-Identifier: MIT

(function_declaration body: (block) @local.scope)
(block) @local.scope

(function_declaration name: (identifier) @local.definition.function)
(parameter name: (identifier) @local.definition.parameter)
(local_var_declaration name: (identifier) @local.definition.var)
(global_var_declaration name: (identifier) @local.definition.var)
(range_loop_statement variable: (identifier) @local.definition.var)

(identifier) @local.reference
