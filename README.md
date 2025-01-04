## Diploma of Munkherdene-B210900003

Энэхүү дипломын ажил нь монгол хэлтэй программчлалын хэлны судалгаа болон бүтээл болно.

#### Components

- [ ] Lexer
- [ ] Parser
- [ ] Analyzer
- [ ] Code Generation
- [ ] Optimization

#### Workflow & Requirements

Do i have to write lexer by myself?
Yes. Because that makes me understand the part of compiler

- [x] Hand Parser Or Generator. Choose?
      Hand

Do i really need low level language?
Some memory level control

Should i use compiler backend for compability of any type of cpu or specific arch?
Maybe

#### Formal grammar

expression = primary;
primary = NUMBER | STRING | IDENT | "(" expression ")";

#### References

Book -- Writing a C Compiler: Build a Real Programming Language from Scratch Paperback - by Nora Sandler

Book -- Compilers: Principles, Techniques, and Tools

Book -- Crafting Interpreters: Nystrom, Robert

Book -- Writing A Compiler In Go: Thorsten Ball
