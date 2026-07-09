# selfhost — mon_lang emitting native code from mon_lang

This directory shows mon_lang producing a **native executable with zero external
tooling**, written in mon_lang itself. No `as`, no `cc`, no `ld` — the program
lays out a Mach-O and writes the machine code byte by byte.

| File           | Role                                                             |
| -------------- | ---------------------------------------------------------------- |
| `макхо.mn`     | a static x86_64 Mach-O writer (`макхо_бич`) — header, load        |
|                | commands, `__TEXT`, entry point                                  |
| `self_emit.mn` | driver: builds `exit(42)` machine code and emits a runnable file |

## Run

Compile and run `self_emit.mn` with the native back end; it writes an
executable to `/tmp/mon_self_emit_out`, which then runs and exits `42`:

```sh
mon_lang gen selfhost/self_emit.mn -o /tmp/emitter && /tmp/emitter
/tmp/mon_self_emit_out; echo $?   # -> 42
```

This is exercised end to end by `tests/selfemit_test.go`.

## Notes

- The canonical self-hosted compiler lives in the `mon` submodule
  (`../mon/компилятор/`).
- The earlier asm-emitting self-host (which produced AT&T `.s` to assemble with
  `cc` against a C runtime) was removed along with the legacy `cc`/`as` link
  path. The compiler now emits native machine code directly on both x86_64 and
  arm64, so a C toolchain is never involved.
