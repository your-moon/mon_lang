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
<program> ::= <function>
<function> ::= "функц" <identifier> "(" ")" "->" "тоо" { <block-item> }
<block> ::= "{" <block-item>* "}"
<block-item> ::= <statement> | <declaration>
<declaration> ::= "зарла" <identifier> [ ":" "тоо" ] [ "=" <exp> ] ";"
<statement> ::=   "буц" <exp> ";"
                | "хэрэв" <exp> "бол" <block> [ "үгүй бол" <block> ]
                | "давтах" <identifier> "=" <exp> "-с" <exp> "хүртэл" <block>
                | <exp> ";"
                | ";"
<exp> ::= <factor>
        | <exp> <binop> <exp>
        | <exp> "?" <exp> ":" <exp>
        | <identifier> "=" <exp>
<factor> ::= <int>
           | <string>
           | <identifier>
           | <unop> <factor>
           | "(" <exp> ")"
<unop> ::= "-" | "~" | "!"
<binop> ::= "-" | "+" | "*" | "/"
          | "болон" | "эсвэл"
          | "==" | "!=" | "<" | "<=" | ">" | ">="
<identifier> ::= ? An identifier token ?
<int> ::= ? A constant token ?
<string> ::= ? A string literal token ?

```

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
