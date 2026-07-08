# Known codegen bug: subtraction-in-guard + call-in-following-loop

Status: **open** — blocks the multistage compiler's parser stage (`mc/парсер`).

## Minimal reproduction

```mon
extern функц мөр_урт(с мөр) -> тоо64 {}
extern функц байт(с мөр, и тоо) -> тоо {}
extern функц хэвлэ(н тоо64) -> хоосон {}

функц кв(эх мөр, э тоо, т тоо, к мөр) -> тоо {
    зарла у: тоо = мөр_урт(к);
    хэрэв (т - э) != у бол { буц 0; }      // guard with a subtraction
    зарла и: тоо = 0;
    давтах и < у бол {
        хэрэв байт(эх, э + и) != байт(к, и) бол { буц 0; }  // loop with calls
        и = и + 1;
    }
    буц 1;
}
функц үндсэн() -> тоо { хэвлэ(кв("функц нэм", 0, 10, "функц")); буц 0; }
```

Expected `1`. Native encoder prints `0`; the `--cc` path **segfaults** (exit 139).

## What triggers it

Both of these are required — removing either makes it pass:

1. the guard compares a **subtraction** (`(т - э) != у`), and
2. the following `давтах` loop contains a **function call** (`байт(...)`).

Adding any intervening statement that spills (e.g. a `хэвлэ`) also masks it.
Isolated (guard alone, or loop alone, or `т != у` without the subtraction, or a
loop with no call) all compile and run correctly.

## What was ruled out

- **Register allocator** (`code_gen/pseudo.go`): setting `calleeSaved = {}`
  (memory-only) still reproduces.
- **Optimizer** (`tackygen/optimize.go`): disabling `Optimize()` still
  reproduces.
- **Structs**: reproduces with plain scalar params, no structs.

## Observations

The emitted AT&T text (`gen --cc --asm`) *looks* correct — the guard computes
`10 - 0 == 10` and does not branch — yet the assembled `--cc` binary segfaults
and the native-encoded binary returns the wrong value. That both back ends
misbehave (differently) from the same `code_gen` AST points at the shared
Tacky→ASM lowering or a mis-encoding surfaced only under this register/stack
shape. Needs instruction-level bisection of the native encoder output vs. the
AT&T text.

## Impact

The parser's `кв` (keyword match) and any similar length-guard-then-scan loop
miscompile, so `mc/парсер` cannot be exercised until this is fixed. The parser
source itself (`mc/парсер/*.mn`) is correct mon_lang.
