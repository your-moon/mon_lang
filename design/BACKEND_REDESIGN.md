# mc backend redesign — a QBE-style SSA backend

The self-hosted `mc` compiler's backend is being rebuilt on a small **typed SSA
IR** with real **liveness** and **linear-scan register allocation**, instead of
an ad-hoc stack machine with a bolt-on allocator. This is the design that makes
the whole class of "value corrupted across a call" bugs impossible by
construction (see `CODEGEN_BUG_guard_sub_loopcall.md` for why the old shape is
fragile).

Inspirations: **QBE** (Quentin Carbonneaux — a ~10k-line SSA backend proving you
don't need LLVM to be correct), **chibicc/TCC** (disciplined value handling in a
tiny compiler), and **Go's SSA** (order-independent by construction).

## Pipeline

```
AST  ──lower──▶  SSA IR  ──▶  liveness  ──▶  linear-scan alloc  ──▶  emit
(mc/парсер)     (mc/ир)                                            (mc/кодген)
                                                             ├─ x86_64  + Mach-O
                                                             └─ arm64   + Mach-O (+ ad-hoc sign)
```

## The IR

A function is a list of **basic blocks**; each block is a sequence of
**instructions** ending in a **terminator**. Values are **temps** (SSA — each
assigned once); an operand is either a temp, an immediate, or a symbol.

Instructions (QBE-flavoured, mon_lang's self-hostable subset — all values are
64-bit for now):

| op        | form                    | meaning                        |
| --------- | ----------------------- | ------------------------------ |
| `КОПИ`    | `%d = копи v`           | copy                           |
| `НЭМ/ХАС/…`| `%d = нэм a, b`        | arithmetic (+ − × ÷ %)         |
| `ХАРЬЦ_*` | `%d = харьц_бн a, b`    | compare → 0/1 (eq/ne/lt/…)     |
| `АЧААЛ`   | `%d = ачаал a`          | load 8 bytes at address a      |
| `ХАДГАЛ`  | `хадгал a, v`           | store v at address a           |
| `ДУУД`    | `%d = дууд fn(args…)`   | call                           |

Terminators: `ҮСРЭ L` (jmp), `САЛАА c, L1, L2` (branch), `БУЦ v` (return).

Blocks are numbered; every temp has exactly one defining instruction, so
liveness is a straightforward backward dataflow (live-in/live-out per block).

## Register allocation

Classic **linear scan** (Poletto & Sarkar): compute live intervals over a linear
ordering of instructions, then scan intervals in start order assigning machine
registers, spilling the interval with the furthest end when registers run out.
Because allocation is driven by *intervals that end*, a value live across a call
is never handed a caller-saved register unless it is spilled — the old bug
cannot recur.

## Status

- [x] IR types + builder (`mc/ир/заавар.mn`)
- [x] AST → IR lowering for expressions (`mc/ир/лоуэр.mn`)
- [x] liveness (`mc/ир/амьдрал.mn`)
- [x] linear-scan allocator (`mc/ир/хуваарь.mn`)
- [ ] lowering: statements + control flow (if / while) + locals (needs a temp
      symbol table)
- [ ] lowering: calls
- [ ] x86_64 emit (register-based, from the allocation) + own Mach-O writer
- [ ] arm64 emit + Mach-O + ad-hoc code signature
- [ ] wire mc's driver to the new backend; reach the self-hosting fixed point

The algorithmic core (IR + liveness + linear-scan) is done and tested. The
remaining work is emission: turning allocated IR into machine bytes for each
target and writing the Mach-O container — the largest block, reusing the
techniques already proven in the Go compiler's `encoder/` and `macho/`.
