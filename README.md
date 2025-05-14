# Монгол Хэлний Компайлер

Монгол хэл дээр суурилсан програмчлалын хэлний компайлер юм. Энэ төсөл нь лексик анализээс эхлүүлж код үүсгэлт хүртэлх компайлер үйл явцыг бүрэн хэрэгжүүлсэн.

## 🚀 Онцлог

- **Монгол Хэлний Синтакс**: Монгол хэлний түлхүүр үг, грамматик ашиглан програмчлалын хэлны загварыг үүсгэсэн.
- **Бүрэн Компайлер Үйл Явц**:
  - Лексик Анализ (Lexer)
  - Синтакс Анализ (Parser)
  - Семантик Анализ
  - Код Үүсгэлт
  - Линклэлт
- **Орчин Үеийн Хэлний Онцлогууд**:
  - Хүчтэй төрлийн систем
  - Функц тодорхойлолт болон дуудалт
  - Удирдлагын урсгалын мэдэгдлүүд (хэрэв, давт, давтах)
  - Хувьсагчийн тодорхойлолт болон утга оноох
  - Импорт систем
  - Хүрээний илэрхийлэл

## 📚 Хэлний Синтакс

```mon
функц нэмэх(а: тоо, б: тоо) -> тоо {
    зарла нийлбэр: тоо = а + б;
    буц нийлбэр;
}

функц үндсэн() {
    зарла тоо: тоо = 5;
    хэрэв тоо > 0 бол {
        зарла үр_дүн = нэмэх(тоо, 10);
    }
}
```

## 🛠️ Архитектур

Компайлер нь уламжлалт үйл явцын архитектурт дагаж мөрддөг:

```
Lexer -> Parser -> Compiler -> Code Generator -> Linker
```

### Бүрэлдэхүүн хэсгүүд

1. **Lexer**: Эх кодыг утга учиртай нэгжүүд болгон задалдаг
2. **Parser**: Токенуудаас Абстракт Синтакс Мод (AST) бүтээдэг
3. **Compiler**: Семантик анализ болон оновчлол хийдэг
4. **Code Generator**: Зорилтот код (x86 assembly) үүсгэдэг
5. **Linker**: Эцсийн гүйцэтгэх файл үүсгэдэг

## 🔧 Суулгах болон Ажиллуулах

```bash
# Компайлер бүтээх
make build

# Компайлер ажиллуулах
./compiler input.mon

# Програм компайл хийж ажиллуулах
./compiler input.mon && ./a.out
```

## 📖 Хэлний Тодорхойлолт

Хэл нь EBNF грамматик дээр суурилдаг бөгөөд дараах зүйлсийг тодорхойлсон:
- Програмын бүтэц
- Функцийн тодорхойлолт
- Хувьсагчийн тодорхойлолт
- Удирдлагын урсгалын мэдэгдлүүд
- Илэрхийлэл болон операторууд
- Төрлийн систем

## 🎯 Төслийн Зорилго

- Монгол хэл дээр суурилсан програмчлалын хэлний бүрэн ажиллах компайлер бүтээх
- Алдааны мэдээлэл болон тайлбарыг зөв хийх
- Үр дүнтэй x86 assembly код үүсгэх
- Монгол хэлт хэрэглэгчдэд цэвэр, ойлгомжтой синтакс санал болгох

## 📚 Лавлагаа

- [Monkey Language Implementation](https://github.com/kitasuke/monkey-go)
- [x86 Assembly Reference](https://flint.cs.yale.edu/cs421/papers/x86-asm/asm.html)
- [Oracle x86 Assembly Manual](https://docs.oracle.com/cd/E19253-01/817-5477/817-5477.pdf)

## 🤝 Хамтын Ажиллагаа

Хамтын ажиллагааг урьж байна! Pull Request илгээхэд бэлэн байна.

## 📝 Лиценз

Энэ төсөл нь MIT лицензээр хамгаалагдсан. Дэлгэрэнгүйг LICENSE файлаас үзнэ үү.

## 👨‍💻 Зохиогч

[Э. Мөнх-Эрдэнэ 2025/Компьютерын ухаан] - Төгсөлтийн Төсөл

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
