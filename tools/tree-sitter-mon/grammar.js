/**
 * tree-sitter grammar for mon_lang — a Mongolian-keyword systems language.
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT
 */

/* Identifiers accept ASCII letters plus the Cyrillic block, matching the
 * compiler's lexer (any byte >= 128 is treated as a letter). */
const IDENT = /[A-Za-z_Ѐ-ӿԀ-ԯ][A-Za-z0-9_Ѐ-ӿԀ-ԯ]*/;

const PREC = {
  or: 1,
  and: 2,
  equality: 3,
  comparison: 4,
  additive: 5,
  multiplicative: 6,
  cast: 7,
  unary: 8,
  postfix: 9,
};

module.exports = grammar({
  name: 'mon',

  word: $ => $.identifier,

  extras: $ => [/\s/, $.line_comment],

  conflicts: $ => [],

  rules: {
    source_file: $ => repeat($._top_level),

    _top_level: $ => choice(
      $.import_declaration,
      $.function_declaration,
      $.extern_declaration,
      $.struct_declaration,
      $.impl_block,
      $.global_var_declaration,
    ),

    // ---- imports: ашигла "path" [гэж name] ; ----
    import_declaration: $ => seq(
      'ашигла',
      field('path', $.string_literal),
      optional(seq('гэж', field('alias', $.identifier))),
      ';',
    ),

    // ---- functions ----
    function_declaration: $ => seq(
      optional('тунх'),
      'функц',
      field('name', $.identifier),
      field('parameters', $.parameter_list),
      '->',
      field('return_type', $._type),
      // block body, explicit ';', or a bare header (forward prototype)
      optional(choice(field('body', $.block), ';')),
    ),

    extern_declaration: $ => seq(
      'extern',
      'функц',
      field('name', $.identifier),
      field('parameters', $.parameter_list),
      '->',
      field('return_type', $._type),
      choice(';', field('body', $.block)),
    ),

    parameter_list: $ => seq(
      '(',
      optional(seq(
        $._param_first,
        repeat(seq(',', $.parameter)),
      )),
      ')',
    ),

    // the receiver `өөрөө` may stand alone as the first parameter
    _param_first: $ => choice($.self_parameter, $.parameter),
    self_parameter: $ => 'өөрөө',

    parameter: $ => seq(
      field('name', $.identifier),
      field('type', $._type),
    ),

    // ---- structs & impls ----
    struct_declaration: $ => seq(
      'бүтэц',
      field('name', $.identifier),
      '{',
      optional(seq(
        $.field_declaration,
        repeat(seq(',', $.field_declaration)),
        optional(','),
      )),
      '}',
    ),

    field_declaration: $ => seq(
      field('name', $.identifier),
      ':',
      field('type', $._type),
    ),

    impl_block: $ => seq(
      'хэрэгжүүл',
      field('type', $.identifier),
      '{',
      repeat($.function_declaration),
      '}',
    ),

    // ---- variable declarations ----
    global_var_declaration: $ => seq(
      optional(choice('тунх', 'статик')),
      $._var_decl_inner,
      ';',
    ),

    local_var_declaration: $ => seq(
      $._var_decl_inner,
      ';',
    ),

    _var_decl_inner: $ => seq(
      'зарла',
      field('name', $.identifier),
      ':',
      field('type', $._type),
      optional(seq('=', field('value', $._expression))),
    ),

    // ---- types ----
    _type: $ => choice(
      $.primitive_type,
      $.identifier,          // struct types
      $.array_type,
      $.pointer_type,
    ),

    array_type: $ => seq(
      choice($.primitive_type, $.identifier),
      '[',
      optional($._expression),
      ']',
    ),

    pointer_type: $ => seq(
      choice($.primitive_type, $.identifier),
      repeat1('*'),
    ),

    primitive_type: $ => choice(
      'тоо', 'тоо64', 'этоо', 'этоо64',
      'тэмдэгт', 'бутархай', 'мөр', 'хоосон',
    ),

    // ---- statements ----
    block: $ => seq('{', repeat($._statement), '}'),

    _statement: $ => choice(
      $.local_var_declaration,
      $.if_statement,
      $.while_statement,
      $.range_loop_statement,
      $.match_statement,
      $.return_statement,
      $.break_statement,
      $.continue_statement,
      $.assignment_statement,
      $.expression_statement,
      $.block,
    ),

    if_statement: $ => prec.right(seq(
      'хэрэв',
      field('condition', $._expression),
      'бол',
      field('consequence', $.block),
      optional(seq(
        'эсвэл',
        field('alternative', choice($.if_statement, $.block)),
      )),
    )),

    while_statement: $ => seq(
      'давтах',
      field('condition', $._expression),
      'бол',
      field('body', $.block),
    ),

    // давт и бол 1..5 хүртэл { ... }
    range_loop_statement: $ => seq(
      'давт',
      field('variable', $.identifier),
      'бол',
      field('start', $._expression),
      '..',
      field('end', $._expression),
      'хүртэл',
      field('body', $.block),
    ),

    match_statement: $ => seq(
      'тааруул',
      field('value', $._expression),
      '{',
      repeat($.match_arm),
      '}',
    ),

    match_arm: $ => seq(
      field('pattern', choice($._expression, $.wildcard_pattern)),
      '=>',
      field('body', $.block),
    ),

    wildcard_pattern: $ => '_',

    return_statement: $ => seq(
      'буц',
      optional(field('value', $._expression)),
      ';',
    ),

    break_statement: $ => seq('зогс', ';'),
    continue_statement: $ => seq('үргэлжлүүл', ';'),

    assignment_statement: $ => seq(
      field('target', $._lvalue),
      '=',
      field('value', $._expression),
      ';',
    ),

    _lvalue: $ => choice(
      $.identifier,
      $.field_expression,
      $.index_expression,
      $.dereference_expression,
    ),

    expression_statement: $ => seq($._expression, ';'),

    // ---- expressions ----
    _expression: $ => choice(
      $.binary_expression,
      $.unary_expression,
      $.dereference_expression,
      $.cast_expression,
      $.call_expression,
      $.method_call_expression,
      $.field_expression,
      $.index_expression,
      $.new_expression,
      $.parenthesized_expression,
      $.identifier,
      $.self_expression,
      $._literal,
    ),

    parenthesized_expression: $ => seq('(', $._expression, ')'),

    self_expression: $ => 'өөрөө',

    binary_expression: $ => {
      const table = [
        ['||', PREC.or],
        ['&&', PREC.and],
        ['==', PREC.equality],
        ['!=', PREC.equality],
        ['<', PREC.comparison],
        ['>', PREC.comparison],
        ['<=', PREC.comparison],
        ['>=', PREC.comparison],
        ['+', PREC.additive],
        ['-', PREC.additive],
        ['*', PREC.multiplicative],
        ['/', PREC.multiplicative],
        ['%', PREC.multiplicative],
      ];
      return choice(...table.map(([op, p]) => prec.left(p, seq(
        field('left', $._expression),
        field('operator', op),
        field('right', $._expression),
      ))));
    },

    unary_expression: $ => prec(PREC.unary, seq(
      // - negate, ! not, ~ bitnot, & address-of
      field('operator', choice('-', '!', '~', '&')),
      field('operand', $._expression),
    )),

    // pointer dereference — valid as both a value and an assignment target
    dereference_expression: $ => prec(PREC.unary, seq(
      '*',
      field('operand', $._expression),
    )),

    cast_expression: $ => prec(PREC.cast, seq(
      field('value', $._expression),
      'гэж',
      field('type', choice($.primitive_type, $.identifier)),
    )),

    call_expression: $ => prec(PREC.postfix, seq(
      field('function', $.identifier),
      field('arguments', $.argument_list),
    )),

    method_call_expression: $ => prec(PREC.postfix, seq(
      field('receiver', $._expression),
      '.',
      field('method', $.identifier),
      field('arguments', $.argument_list),
    )),

    field_expression: $ => prec(PREC.postfix, seq(
      field('object', $._expression),
      '.',
      field('field', $.identifier),
    )),

    index_expression: $ => prec(PREC.postfix, seq(
      field('array', $._expression),
      '[',
      field('index', $._expression),
      ']',
    )),

    new_expression: $ => prec(PREC.unary, seq(
      'шинэ',
      field('type', choice($.primitive_type, $.identifier)),
      '[',
      field('size', $._expression),
      ']',
    )),

    argument_list: $ => seq(
      '(',
      optional(seq($._expression, repeat(seq(',', $._expression)))),
      ')',
    ),

    // ---- literals ----
    _literal: $ => choice(
      $.number_literal,
      $.float_literal,
      $.char_literal,
      $.string_literal,
    ),

    number_literal: $ => /\d+/,
    float_literal: $ => /\d+\.\d+/,

    char_literal: $ => seq(
      "'",
      choice(/[^'\\]/, $.escape_sequence),
      "'",
    ),

    string_literal: $ => seq(
      '"',
      repeat(choice(
        token.immediate(/[^"\\]+/),
        $.escape_sequence,
      )),
      '"',
    ),

    escape_sequence: $ => token.immediate(/\\["'\\nrt0]/),

    identifier: $ => IDENT,

    line_comment: $ => token(seq('//', /.*/)),
  },
});
