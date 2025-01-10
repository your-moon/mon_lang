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

expression = primary;
primary = NUMBER | STRING | IDENT | "(" expression ")";

- newline -> \r, \n, \r\n

- а -> 0x0430
- б -> 0x0431
- в -> 0x0432
- г -> 0x0433
- д -> 0x0434
- е -> 0x0435
- ж -> 0x0436
- з -> 0x0437
- и -> 0x0438
- й -> 0x0439
- к -> 0x043a
- л -> 0x043b
- м -> 0x043c
- н -> 0x043d
- о -> 0x043e
- п -> 0x043f
- р -> 0x0440
- с -> 0x0441
- т -> 0x0442
- у -> 0x0443
- ф -> 0x0444
- х -> 0x0445
- ц -> 0x0446
- ч -> 0x0447
- ш -> 0x0448
- щ -> 0x0449
- ъ -> 0x044a
- ы -> 0x044b
- ь -> 0x044c
- э -> 0x044d
- ю -> 0x044e
- я -> 0x044f
- (space) -> 0x0020
- ё -> 0x0451
- ү -> 0x04af
- ө -> 0x04e9

#### References

Book -- Writing a C Compiler: Build a Real Programming Language from Scratch Paperback - by Nora Sandler

Book -- Compilers: Principles, Techniques, and Tools

Book -- Crafting Interpreters: Nystrom, Robert

Book -- Writing A Compiler In Go: Thorsten Ball
