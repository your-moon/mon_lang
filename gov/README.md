#### Architecture

Lexer -> Parser -> Compiler -> Gen -> Link

#### References

- https://github.com/kitasuke/monkey-go
- https://flint.cs.yale.edu/cs421/papers/x86-asm/asm.html
- https://docs.oracle.com/cd/E19253-01/817-5477/817-5477.pdf

#### Problems

- parsing the expression left recursive way

```sh
===EBNF===

(* Program Structure *)
<program> ::= { <import> } <function>
<import> ::= "импорт" <identifier> { "." <identifier> } ";"
<function> ::= "функц" <identifier> "(" [ <param-list> ] ")" [ "->" <type> ] <block>
<block> ::= "{" <block-item>* "}"
<block-item> ::= <statement> | <declaration>

(* Declarations *)
<declaration> ::= <fn-decl> | <var-decl>
<var-decl> ::= "зарла" <identifier> [ ":" <type> ] [ "=" <exp> ] ";"
<fn-decl> ::= <identifier> "(" [ <param-list> ] ")" [ "->" <type> ] ";"
<param-list> ::= <param> { "," <param> }
<param> ::= <identifier> ":" <type>
<type> ::= "тоо" | "тэмдэгт"

(* Statements *)
<statement> ::= <return-stmt>
              | <if-stmt>
              | <for-stmt>
              | <while-stmt>
              | <break-stmt>
              | <continue-stmt>
              | <exp-stmt>
              | ";"

<return-stmt> ::= "буц" <exp> ";"
<if-stmt> ::= "хэрэв" <exp> "бол" <block> [ "үгүй бол" <block> ]
<for-stmt> ::= "давт" <identifier> "бол" <exp> "хүртэл" <block>
<while-stmt> ::= "давтах" [ <exp> ] "хүртэл" <block>
<break-stmt> ::= "зогс" ";"
<continue-stmt> ::= "үргэлжлүүл" ";"
<exp-stmt> ::= <exp> ";"

(* Expressions *)
<exp> ::= <assignment>
<assignment> ::= <identifier> "=" <exp>
               | <logical-or>
<logical-or> ::= <logical-and> { "эсвэл" <logical-and> }
<logical-and> ::= <equality> { "болон" <equality> }
<equality> ::= <comparison> { ("==" | "!=") <comparison> }
<comparison> ::= <term> { ("<" | "<=" | ">" | ">=") <term> }
<term> ::= <factor> { ("+" | "-") <factor> }
<factor> ::= <unary> { ("*" | "/") <unary> }
<unary> ::= ("-" | "~" | "!") <unary> | <primary>
<primary> ::= <number>
            | <string>
            | <identifier>
            | <fn-call>
            | <range-exp>
            | "(" <exp> ")"

<fn-call> ::= <identifier> "(" [ <argument-list> ] ")"
<argument-list> ::= <exp> { "," <exp> }
<range-exp> ::= <exp> ".." <exp>

(* Lexical Elements *)
<identifier> ::= <letter> { <letter> | <digit> }
<number> ::= <digit> { <digit> }
<string> ::= '"' { <char> } '"'
<letter> ::= "A" | "B" | ... | "Z" | "a" | "b" | ... | "z" | "_"
<digit> ::= "0" | "1" | ... | "9"
<char> ::= (* Any character except unescaped double quote *)

(* Comments *)
<comment> ::= "//" { <char> } <newline>

```sh
===AST===

program = Program(function_definition)
function_definition = Function(identifier name, parameter* params, type return_type, block_item* body)
parameter = Parameter(identifier name, type type)
type = Number | String
block_item = S(statement) | D(declaration)
declaration = Declaration(identifier name, type? type, exp? init)
statement = Return(exp)
         | If(exp, block, block?)
         | For(identifier, exp, exp, block)
         | Expression(exp)
         | Null
block = Block(block_item* items)
exp = Constant(int)
    | String(string)
    | Var(identifier)
    | ArrayAccess(identifier, exp)
    | Unary(unary_operator, exp)
    | Binary(binary_operator, exp, exp)
    | Assignment(exp, exp)
    | Ternary(exp, exp, exp)
unary_operator = Complement | Negate | Not
binary_operator = Add | Subtract | Multiply | Divide | Remainder
                | And | Or
                | Equal | NotEqual | LessThan | LessOrEqual
                | GreaterThan | GreaterOrEqual


```
