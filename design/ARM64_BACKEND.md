# Native arm64 (Apple Silicon) backend

`gen --arch arm64` lowers the architecture-neutral **Tacky IR** straight to a
native, self-signed arm64 Mach-O executable with **no external tooling** — no
cc/as/ld and no `codesign`. The whole toolchain is in-tree.

```
Tacky IR ──armgen──▶ arm64 machine code ─┐
   (arch-neutral)     armgen/gen.go       ├─▶ macho.WriteExecutableARM64
                      armgen/asm.go        │      (signed Mach-O)
                      armgen/stdlib.go ────┘
```

## Why it's more than the x86 writer

Modern macOS (Darwin 25) will not exec a pure static `LC_UNIXTHREAD` arm64
binary, and it SIGKILLs any unsigned arm64 binary. So, unlike the x86 path
(which runs under Rosetta), the arm64 writer must:

1. Be **dyld-loaded**: `LC_MAIN` + `LC_LOAD_DYLINKER` + `LC_LOAD_DYLIB`
   (libSystem) + empty `LC_DYLD_CHAINED_FIXUPS`. All I/O is still raw `svc`
   syscalls; we import nothing.
2. Carry its own **ad-hoc code signature** that satisfies runtime AMFI (much
   stricter than `codesign -v`): a CodeDirectory over 16 KiB page hashes plus
   Requirements and (empty) signature blobs. Derisked byte-for-byte against
   `codesign`'s output — see the "arm64 Mach-O recipe" memory note.

## Codegen model

A **template / stack-machine** codegen (`armgen/gen.go`): every Tacky temp gets
an 8-byte frame slot; each instruction loads operands into x8/x9(/x10), operates,
and stores back. Correct by construction — no register allocator, so the whole
class of "value corrupted across a call" bugs cannot occur (x8/x9 never span a
call; args go in x0–x7). Structs and arrays are heap-allocated by the front end;
`monAlloc` is a bump allocator over `mmap`'d chunks.

Pointer loads/stores are **width-aware**: per-temp byte widths recovered from the
existing asm-symbol pass drive 4- vs 8-byte `ldr/str`, so sub-word (`тоо`/Int32)
struct fields aren't read or written 8 bytes wide.

## Status

Coverage (`MON_TEST_ARM64=1`, arm64 host): **33 run/ tests pass, 0 failures.**
Working: integer/pointer arithmetic, comparisons, if/while, locals, recursion +
calls, globals (`__DATA`), strings, UTF-8 char output, unsigned ops, sized
arrays, structs + methods, the syscall stdlib, and the bump heap.

Not yet supported (skip): doubles/floats (`IntToDouble`), `файл_бичих` file
write, `унш` stdin parse, and two liveness self-host stress tests
(`амьдрал/хуваарь_туршилт`) that hit an elusive codegen edge case — every
reduced reproducer (structs, struct arrays, methods, in-place field writes,
array params) passes individually, so it remains under investigation.

The x86_64 path is untouched and fully green.
