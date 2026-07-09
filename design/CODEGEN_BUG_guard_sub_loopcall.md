# Codegen bugs: register allocator (FIXED) + guard-before-loop (worked around)

Status: two distinct bugs.

1. **Register allocator corrupts values across method calls — FIXED.** The
   "linear-scan-lite" allocator in `code_gen/pseudo.go` assigned callee-saved
   registers without liveness analysis, so a value still live across a
   `өөрөө.method(arg)` call in a loop got clobbered (segfault / wrong bytes).
   This was pervasive — every struct method that calls another in a loop was
   affected. **Fix:** empty `calleeSaved` (disable the allocator; everything
   lives on the stack). Full suite green, self-hosting intact, and the method
   reproducers below now return correct results. A principled replacement is
   `mc/ир` (SSA + real liveness + linear-scan; see BACKEND_REDESIGN.md).

2. **Guard-with-subtraction before a call-loop — still open, worked around.**
   The scalar case below still miscompiles even with the allocator disabled, so
   `mc/парсер`'s `кв` moves its length check *after* the byte loop. Documented
   here for a proper fix.

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

## Working around it

`mc/парсер/парсер.mn`'s `кв` avoids the pattern by moving the length check
**after** the byte-compare loop (nothing with a call follows the comparison):

```mon
давтах и < у бол {
    хэрэв э + и >= т бол { буц 0; }
    хэрэв байт(өөрөө.у.эх, э + и) != байт(к, и) бол { буц 0; }
    и = и + 1;
}
хэрэв (т - э) != у бол { буц 0; }   // length check AFTER the loop
буц 1;
```

This compiles and runs correctly on both back ends. The underlying bug remains.

## What was ruled out

- **Register allocator** (`code_gen/pseudo.go`): setting `calleeSaved = {}`
  (memory-only) still reproduces (with a *different* symptom — the guard
  misfires instead of the loop).
- **Optimizer** (`tackygen/optimize.go`): disabling `Optimize()` still
  reproduces.
- **Structs**: reproduces with plain scalar params, no structs.
- **Builtins clobbering callee-saved regs**: `_bayt` / `_mqr_urt` in
  `encoder/stdlib.go` only touch caller-saved registers (RAX/RCX/RDX/RSI/RDI/
  R10); base-codegen scratch is R10/R11 (`code_gen/emitter.go`). So the loop's
  calls do preserve `%rbx`/`%r12`-`%r15`.

## The remaining mystery

Every emitted AT&T variant (`--cc --asm`), with or without regalloc, *reads as
correct* — the guard computes `10 - 0 == 10` and does not branch, `к`'s slot is
written once and never overwritten — yet the native binary returns the wrong
value and the `--cc` binary segfaults. Because **both** back ends misbehave from
the same `code_gen` AST, either (a) the shared Tacky→ASM lowering emits a
correct-looking but semantically wrong sequence, or (b) there is a slot/stack
inconsistency that both final stages inherit. Next: dump the **Tacky IR** for
`кв` and diff the temp/var slots, and disassemble the native `_kv` bytes to
compare against the `--asm` text instruction-for-instruction.

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
