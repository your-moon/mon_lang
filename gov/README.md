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
<function> ::= "функц" <identifier> "(" "" ")" "->" "тоо" "{" { <block-item> } "}"
<block-item> ::= <statement> | <declaration>
<declaration> ::= "зарла" <identifier> [ ":" "тоo" ] "=" <exp> ";"
<statement> ::= "буц" <exp> ";" | <exp> ";" | ";"
<exp> ::= <factor> | <exp> <binop> <exp>
<factor> ::= <int> | <identifier> | <unop> <factor> | "(" <exp> ")"
<unop> ::= "-" | "~" | "!"
<binop> ::= "-" | "+" | "*" | "/" | "%" | "&&" | "||"
 | "==" | "!=" | "<" | "<=" | ">" | ">=" | "="
<identifier> ::= ? An identifier token ?
<int> ::= ? A constant token ?

```

```sh
===AST===

program = Program(function_definition)
function_definition = Function(identifier name, block_item* body)
block_item = S(statement) | D(declaration)
declaration = Declaration(identifier name, exp? init)
statement = Return(exp) | Expression(exp) | Null
exp = Constant(int)
 | Var(identifier)
 | Unary(unary_operator, exp)
 | Binary(binary_operator, exp, exp)
 | Assignment(exp, exp)
unary_operator = Complement | Negate | Not
binary_operator = Add | Subtract | Multiply | Divide | Remainder | And | Or
 | Equal | NotEqual | LessThan | LessOrEqual
 | GreaterThan | GreaterOrEqual


```
