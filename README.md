<div align="center">
  <h1>Монгол Хэлний Компайлер</h1>
  <p>Монгол хэл дээр суурилсан програмчлалын хэлний компайлер</p>
  <p>
    <a href="#онцлог">Онцлог</a> •
    <a href="#хэлний-синтакс">Хэлний Синтакс</a> •
    <a href="#архитектур">Архитектур</a> •
    <a href="#төслийн-бүтэц">Төслийн Бүтэц</a> •
    <a href="#суулгах-болон-ажиллуулах">Суулгах</a>
  </p>
  <p>
    <img src="https://img.shields.io/badge/Version-1.0.0-blue.svg" alt="Version">
    <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License">
    <img src="https://img.shields.io/badge/Status-Development-yellow.svg" alt="Status">
    <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go" alt="Go">
    <img src="https://img.shields.io/badge/From%20Scratch-100%25-orange" alt="From Scratch">
  </p>
</div>

---

<div align="center">
  <h3>👨‍💻 Зохиогч</h3>
  <p><b>Э. Мөнх-Эрдэнэ</b></p>
  <p>2025 оны Төгсөлтийн Төсөл</p>
  <p>Компьютерын Ухааны Тэнхим</p>
</div>

---

<div align="center">
  <p><b>⚠️ Энэ төсөл нь GPT эсвэл бусад AI хэрэгслүүд ашиглалгүйгээр, гараар бүрэн бичигдсэн.</b></p>
  <p>Бүх код, грамматик, архитектур болон дизайны шийдвэрүүд зохиогчийн өөрийн бүтээл юм.</p>
</div>

---

Монгол хэл дээр суурилсан програмчлалын хэлний компайлер юм. Энэ төсөл нь лексик анализээс эхлүүлж код үүсгэлт хүртэлх компайлер үйл явцыг бүрэн хэрэгжүүлсэн.

## 🚀 Онцлог

- **Монгол Хэлний Синтакс**: Монгол хэлний түлхүүр үг, грамматик ашиглан програмчлалын хэлны загварыг үүсгэсэн
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

## 📁 Төслийн Бүтэц

```
.
├── base/               # Үндсэн утгууд болон төрлүүд
├── cli/               # Команд мөрийн интерфэйс
├── code_gen/          # Код үүсгэгч
├── errors/            # Алдааны удирдлага
├── examples/          # Жишээ программууд
├── lexer/            # Лексик анализатор
├── parser/           # Синтакс анализатор
├── semantic_analysis/ # Семантик анализ
├── stdlib/           # Стандарт сан
├── test/             # Тест файлууд
├── util/             # Туслах функцууд
├── main.go           # Үндсэн програм
└── go.mod            # Go модулийн тохиргоо
```

### Төслийн Бүрэлдэхүүн Хэсгүүдийн Тайлбар

1. **Лексик Анализ** (`lexer/`)
   - Эх кодыг токен болгон задалдаг
   - Монгол түлхүүр үгсийг таних
   - Алдааны мэдээллийг бүрдүүлэх

2. **Синтакс Анализ** (`parser/`)
   - Токенуудаас AST бүтээдэг
   - Грамматик дүрмүүдийг шалгадаг
   - Илэрхийллийн бүтцийг тодорхойлдог

3. **Семантик Анализ** (`semantic_analysis/`)
   - Төрлийн шалгалт хийх
   - Хувьсагчийн хамрах хүрээг шалгах
   - Функцийн дуудалтыг шалгах

4. **Код Үүсгэлт** (`code_gen/`)
   - AST-ээс x86 assembly код үүсгэх
   - Регистр хуваарилалт
   - Оновчлол хийх

5. **Стандарт Сан** (`stdlib/`)
   - Үндсэн функцууд
   - Системийн дуудалтууд
   - Туслах функцууд

6. **Алдааны Удирдлага** (`errors/`)
   - Алдааны төрлүүд
   - Алдааны мэдээлэл
   - Алдааны байршлыг тодорхойлох

## 🔧 Суулгах болон Ажиллуулах

### Шаардлага

- Go 1.21 эсвэл түүнээс дээш хувилбар
- GNU Assembler (as)

### Суулгах

```bash
# Төслийг татах
git clone https://github.com/yourusername/mongolian-compiler.git
cd mongolian-compiler

# Хамаарлуудыг суулгах
go mod download

# Компайлер бүтээх
go build -o compiler main.go
```

### Ажиллуулах

```bash
# Жишээ програм компайл хийх
./compiler examples/hello.mon

# Үүссэн assembly файлыг ажиллуулах
as -o out.o out.asm
ld -o program out.o
./program
```

## 📖 Хэлний Тодорхойлолт

Хэл нь EBNF грамматик дээр суурилдаг бөгөөд дараах зүйлсийг тодорхойлсон:
- Програмын бүтэц
- Функцийн тодорхойлолт
- Хувьсагчийн тодорхойлолт
- Удирдлагын урсгалын мэдэгдлүүд
- Илэрхийлэл болон операторууд
- Төрлийн систем

### EBNF Грамматик

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
```

### AST Бүтэц

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

## 🎯 Төслийн Зорилго

- Монгол хэл дээр суурилсан програмчлалын хэлний бүрэн ажиллах компайлер бүтээх
- Алдааны мэдээлэл болон тайлбарыг зөв хийх
- Үр дүнтэй x86 assembly код үүсгэх
- Монгол хэлт хэрэглэгчдэд цэвэр, ойлгомжтой синтакс санал болгох

## 📚 Лавлагаа

- [Monkey Language Implementation](https://github.com/kitasuke/monkey-go)
- [x86 Assembly Reference](https://flint.cs.yale.edu/cs421/papers/x86-asm/asm.html)
- [Oracle x86 Assembly Manual](https://docs.oracle.com/cd/E19253-01/817-5477/817-5477.pdf)

## 📝 Лиценз

Энэ төсөл нь MIT лицензээр хамгаалагдсан. Дэлгэрэнгүйг LICENSE файлаас үзнэ үү.

---

<div align="center">
  <sub>Built with ❤️ by Э. Мөнх-Эрдэнэ</sub>
</div>
