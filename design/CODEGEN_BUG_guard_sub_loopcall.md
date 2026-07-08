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

A **comparison statement placed immediately before a loop that contains function
calls** corrupts a value the loop relies on. Both required:

1. a guard comparison before the loop (`хэрэв разн != у бол { ... }`), and
2. the following `давтах` loop contains a **function call** (`байт(...)`).

Narrowed reproduction (values are provably correct, yet the loop misbehaves):

```mon
функц кв(эх мөр, э тоо, т тоо, к мөр) -> тоо {
    зарла у: тоо = мөр_урт(к);        // 10
    зарла разн: тоо = т - э;          // 10
    // Returning `разн * 1000 + у` here yields 10010 — both values correct.
    хэрэв разн != у бол { буц 111; }  // 10 != 10 is false → must NOT fire (it doesn't)
    зарла и: тоо = 0;
    давтах и < у бол {
        хэрэв байт(эх, э + и) != байт(к, и) бол { буц 999; }  // WRONGLY fires
        и = и + 1;
    }
    буц 222;
}
// кв("функц нэм", 0, 10, "функц") returns 999; without the guard line it
// completes the loop and returns 222.
```

The guard itself evaluates correctly (does not return 111), but its mere
presence makes the subsequent byte comparison in the loop report a mismatch on
bytes that are equal. Removing the guard, or the call inside the loop, fixes it.
Any intervening spill (e.g. a `хэвлэ`) also masks it.

Strongly suggests a **frame-slot allocation / liveness bug** in
`code_gen` — a temporary from the guard aliases a stack slot that is still live
across the loop's calls.

## Precise diagnosis

Instrumenting the loop (return `и` and the two byte values on mismatch) shows
the mismatch fires on the **first** iteration (`и = 0`):

- `байт(эх, э + и)` = `209` (correct: first byte of `ф` in "функц нэм")
- `байт(к, и)`      = `207` (WRONG: should also be `209`)

So parameter **`к` is corrupted** by the time the loop runs — its byte 0 reads
`0xCF` (207) instead of `0xD1` (209), a value not present in "функц" at all, i.e.
`к` no longer points where it should. The corruption only happens when the
guard comparison precedes the loop. Conclusion: a temporary introduced by the
guard is assigned a stack slot that overlaps the still-live parameter `к`
(or otherwise clobbers it) across the loop's calls.

Next step: dump `AsmProgram` before/after `ReplacePseudosInProgram`
(`code_gen/pseudo.go`) for `кв` and check which pseudo shares `к`'s offset —
the slot allocator (`ReplaceOperand` / `CurrentOffset`) is over-reusing an
offset that is still live.

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
