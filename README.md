## Diploma of Munkherdene-B210900003

Энэхүү дипломын ажил нь монгол хэлтэй программчлалын хэлны судалгаа болон бүтээл болно.

#### Components

- [ ] Lexer
- [ ] Parser
- [ ] Analyzer
- [ ] Code Generation
- [ ] Optimization

#### Resources

- https://cplusplus.com/reference/cwchar/wchar_t/
- https://craftinginterpreters.com/appendix-i.html
- https://www.chromium.org/chromium-os/developer-library/reference/linux-constants/syscalls/
- https://www.tutorialspoint.com/assembly_programming/assembly_memory_segments.htm

#### For Test

- https://github.com/sustrik/libmill

#### Workflow & Requirements

Do i have to write lexer by myself?
Yes. Because that makes me understand the part of compiler

- [x] Hand Parser Or Generator. Choose?
      Hand

Do i really need low level language?
Some memory level control

Should i use compiler backend for compability of any type of cpu or specific arch?
Maybe

Cyrillic to en ?

#### Formal grammar

```sh

# EBNF


program = {statement}


return_statement = "буц" expression ";"

variable_declaration = "зарл" identifier ":" type ["=" expression] ";"
type = int_type | "тэмдэгт" | identifier
int_type = "этоо" | "этоо8" | "этоо16" | "этоо32" | "этоо64" | "этоо128"
| "тоо" | "тоо8" | "тоо16" | "тоо32" | "тоо64" | "тоо128"


expression = equality
equality = comparison {("==" | "!=") comparison}
comparison = term {("<" | "<=" | ">" | ">=") term}
term = factor {("+" | "-") factor}
factor = unary {("*" | "/") unary}
unary = ("-" | "!") unary | primary

primary = digit | string | boolean | identifier | "(" expression ")" | "хоосон"
string = '"' {printable} '"' ;
printable = [0x20-0x7E0x0430-0x044F]
boolean = "үнэн" | "худал" ;
identifier = alpha {alpha | digit} ;
digit = [0-9] ;
alpha = [а-яА-ЯҮүЁёӨөa-zA-Z] ;

```

- newline -> \r, \n, \r\n

```sh
- а -> 0x0430 -> a
- б -> 0x0431 -> b
- в -> 0x0432 -> v
- г -> 0x0433 -> g
- д -> 0x0434 -> d
- е -> 0x0435 -> ye
- ж -> 0x0436 -> j
- з -> 0x0437 -> z
- и -> 0x0438 -> i
- й -> 0x0439 -> hi
- к -> 0x043a -> k
- л -> 0x043b -> l
- м -> 0x043c -> m
- н -> 0x043d -> n
- о -> 0x043e -> o
- п -> 0x043f -> p
- р -> 0x0440 -> r
- с -> 0x0441 -> s
- т -> 0x0442 -> t
- у -> 0x0443 -> u
- ф -> 0x0444 -> f
- х -> 0x0445 -> h
- ц -> 0x0446 -> ts
- ч -> 0x0447 -> ch
- ш -> 0x0448 -> sh
- щ -> 0x0449 -> shc
- ъ -> 0x044a -> qi
- ы -> 0x044b -> yi
- ь -> 0x044c -> zi
- э -> 0x044d -> e
- ю -> 0x044e -> yu
- я -> 0x044f -> ya
- (space) -> 0x0020
- ё -> 0x0451 -> yo
- ү -> 0x04af -> w
- ө -> 0x04e9 -> q
```

#### References

Book -- Writing a C Compiler: Build a Real Programming Language from Scratch Paperback - by Nora Sandler

Book -- Compilers: Principles, Techniques, and Tools

Book -- Crafting Interpreters: Nystrom, Robert

Book -- Writing A Compiler In Go: Thorsten Ball
