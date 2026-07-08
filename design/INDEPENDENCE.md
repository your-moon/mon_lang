# Toolchain Independence Design

Goal: `mon_lang build x.mon` produces a runnable executable with **zero external
tools** — no `as`, no `cc`, no Xcode Command Line Tools, no libc. One Go binary
is the whole toolchain (TCC model: tccmacho.c writes executables directly).

Proven feasible on this machine: a hand-built 4 KB static x86_64 Mach-O
(LC_UNIXTHREAD entry, raw syscalls, unsigned) runs on macOS 26 under Rosetta.

## Architecture

```
tackygen IR ──► code_gen (asm AST, existing passes) ──► encoder (bytes) ──► macho (executable file)
                                        └──► x86x64.go text emitter kept for -S / debugging
```

New packages:

### encoder/  (asm AST → x86_64 machine code)
- Input: the closed post-fixup instruction set from code_gen:
  mov l/q, movslq, add/sub/imul l/q, cmp, idiv, cdq/cqo, set<cc>, j<cc>, jmp,
  call, push, ret(+epilogue), sub rsp, lea label(%rip), mov via (%r10),
  not/neg. Operands: reg, imm, disp(%rbp), label(%rip).
- Labels/calls: TCC gsym-style backpatching — emit rel32 placeholder, record
  site, patch when label address known. One flat code buffer for the whole
  program (entry stub + all functions + stdlib), so every call/jump is
  internal rel32. No relocations, no symbol table needed in output.
- REX/ModRM assembly is mechanical for this subset (~15 instruction kinds).

### macho/  (bytes → executable)
- Static Mach-O executable: __PAGEZERO, __TEXT(__text,__cstring),
  __DATA(__data), LC_UNIXTHREAD with rip = entry stub.
- Entry stub: call _wndsen; mov %eax,%edi; mov $0x2000001,%rax; syscall.
- No dyld, no LC_LOAD_DYLIB, no signature (x86_64/Rosetta exempt).
- ARM64 phase 2 will add ad-hoc CodeDirectory signing (SHA-256 page hashes,
  as tccmacho.c does) since arm64 requires it.

### stdlib rewrite  (drop stdlib/lib.c and its cc dependency)
Reimplement the 9 libc-backed functions on raw darwin syscalls
(class 0x2000000): khevle (itoa+write), mqr_khevlekh (strlen+write),
unsh/unsh32 (read+parse), odoo (gettimeofday), khwleekh (select-as-sleep),
delgetsTseverlekh (write escape codes), sanamsargwyToo (xorshift seeded from
time), chqlqqlqkh (no-op until allocator exists). Expressed as asm-AST
builders in Go so the one encoder handles them; appended to every program.

## Phases
1. x86_64 encoder + macho writer + syscall stdlib; linker.go default path
   becomes internal, `--cc` flag keeps old as/cc path during transition.
   Runs under Rosetta as today, but with zero install requirements.
2. ARM64 backend (arm64-gen.c is 2,353 lines — scope proof) + ad-hoc signing
   → native Apple Silicon, Rosetta gone.
3. Linux ELF writer (simpler: no signing, static ELF trivially allowed).

## Compatibility
- integration_test.go must pass unchanged on the internal path.
- `-S` keeps emitting AT&T text via x86x64.go.
- Playground: server stops needing Xcode CLT in its image.
