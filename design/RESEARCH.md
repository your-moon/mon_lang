# RESEARCH: Solved problems mon_lang is about to hit

Research into how real compilers handled the five problems mon_lang faces on its way
to a self-hosted, garbage-collected, Go-like language. mon_lang today: Go-implemented,
Tacky IR (Sandler "Writing a C Compiler" lineage), self-contained x86_64 Mach-O emitter
(own encoder + linker + syscall stdlib, no `as`/`cc`/libc), integer-only, file imports
via `ашигла` with `тунх` public visibility.

Each section: what 2-3 real compilers did, the tradeoffs, and a concrete recommendation.

---

## 1. Self-hosting bootstrap paths

**Go (C → Go, the 1.5 story).** Go's compiler and runtime were C until 2015. Rather
than rewrite by hand, Russ Cox built a purpose-built C→Go translator (`c2go`) that
exploited the fact that a compiler barely uses the hard-to-port C features — no macros,
unions, or pointer arithmetic in hot paths. The 5-phase plan translated the existing C
compiler mechanically into Go-syntax-but-still-C-shaped code, then refined it. From Go
1.5 (Aug 2015) the toolchain needs only a *previous Go binary* as seed; each release is
built by the prior one. The translation was not 100% automatic — some code was
hand-finished — but it removed the C toolchain and got working Go code fast, deferring
the "make it idiomatic" cleanup to later releases.
- https://www.infoq.com/news/2015/01/golang-15-bootstrapped/
- https://dev.to/mrsa1/understanding-bootstrapping-how-gos-compiler-is-written-in-go-5ann

**Rust (OCaml → Rust rewrite).** The first Rust compiler, `rustboot`, was written in
OCaml — chosen because OCaml's algebraic types + pattern matching fit type-system
experimentation. This was a *hand rewrite*, not translation: the team gradually
reimplemented the compiler in Rust, and by 2011 `rustc` compiled itself and the OCaml
dependency was deleted (commit 6997adf7, May 2011). Since then the "stage 0" seed is just
a stored binary from a past commit; each release compiles with the previous stable.
- https://rustc-dev-guide.rust-lang.org/building/bootstrapping/what-bootstrapping-does.html
- https://github.com/iohub/rustboot

**Zig (wasm blob strategy).** Zig's problem was bootstrapping *without* shipping a large
C++ binary or requiring an old Zig. Solution: the self-hosted compiler is compiled (via
LLVM's C backend, all other backends off) to a `wasm32-wasi` ReleaseSmall blob — 2.6 MiB,
`wasm-opt`'d to 2.4 MiB, zstd-compressed to **637 KB** committed to source control. A
purpose-built ~4,000-line `wasm2c` (only the WASI calls the compiler actually makes, no
general-purpose wasm support, no C++ dep) translates the blob to C, the system C compiler
builds a stage2, and stage2 builds Zig from source thereafter. The trick: a tiny frozen
blob + tiny translator replaces a heavyweight build dependency.
- https://ziglang.org/news/goodbye-cpp/
- https://github.com/ziglang/zig/pull/13560

**Tradeoffs.**
- *Mechanical translation (Go)*: fastest path to "it runs in the new language," but
  output is non-idiomatic and needs a long cleanup tail. Needs a translator you must
  write and trust.
- *Hand rewrite (Rust)*: idiomatic result, doubles as a language stress-test, but slow
  and you maintain two compilers until crossover.
- *Frozen blob (Zig)*: solves *reproducible from-source* bootstrap and dependency
  minimization, but only helps *after* you already have a self-hosted compiler — it's a
  distribution/bootstrap-chain solution, not a "how do I get self-hosted" solution.

**Recommendation for mon_lang.** Do a **hand rewrite, Rust-style**, but keep the Go
compiler as the permanent stage-0 seed (Go-implemented compilers make excellent frozen
seeds — this is exactly Go/Rust's model). Concrete smallest realistic path for a teaching
language:
1. First make the Go compiler emit enough of the language to express a compiler:
   strings, arrays, structs, pointers, file I/O, basic maps. (Tasks #8/#6 already target
   this.) This is the real gate — you cannot self-host until the language can express its
   own data structures.
2. Rewrite the compiler *in mon_lang*, front-to-back, keeping the Go version as reference
   oracle: compile the same inputs through both and diff the emitted Mach-O / assembly.
3. Crossover: once `mon-compiler.mn` compiled by the Go compiler can compile
   `mon-compiler.mn` and produce a byte-identical (or behaviorally identical) compiler,
   you are self-hosted. Freeze that Go-built binary as the seed.
Do **not** build a C→mon translator (you have no C to translate — you're Go). Do **not**
adopt the wasm blob until *after* self-hosting, when reproducible bootstrap matters. The
Go seed binary is your `stage0`; check its provenance into the repo/release notes.

---

## 2. Garbage collection without a heavy runtime

**Boehm–Demers–Weiser (conservative mark-sweep).** The canonical "GC for uncooperative
environments." No compiler cooperation needed: it treats every word-aligned value in
stacks, registers, and global/static data as a *potential* pointer, and marks any heap
object whose address appears. Four phases: clear mark bits → mark reachable from roots →
(optionally) finalize → sweep unmarked blocks back to free lists. Drop-in `malloc`
replacement. Downside: false positives (an integer that looks like a heap address pins an
object), no compaction, and it must conservatively scan non-pointer data.
- https://hboehm.info/gc/
- https://en.wikipedia.org/wiki/Boehm_garbage_collector
- https://www.hboehm.info/gc/gcdescr.html

**Go's early GC (start simple).** Go 1.0 (2012) shipped a **stop-the-world, mark-sweep**
collector — halt everything, mark, sweep, resume. Noticeable pauses, but *correct and
simple*, and it shipped. Concurrency (tri-color, concurrent mark-sweep, write barriers)
came only in Go 1.5 (2015) once the language was established. The lesson: STW mark-sweep
was good enough to launch a production language; concurrency was a later optimization.
- https://dev.to/siashish/understanding-gos-garbage-collector-a-detailed-guide-kj4

**TinyGo (a menu of simple options).** TinyGo, targeting constrained devices, offers
exactly the tiers a small language should consider: `leaking` (bump-allocate, never free —
the simplest possible allocator, great for short-lived programs), `conservative` (simple
Boehm-style mark-sweep, works everywhere, any allocation may trigger a full cycle),
`precise` (uses type info to scan only real pointers — added as the wasm default in
v0.34.0 to avoid conservative false positives), and `none`.
- https://tinygo.org/docs/reference/usage/important-options/
- https://aykevl.nl/2020/09/gc-tinygo/

**Lua (tri-color, single-threaded, proven-simple).** Lua uses incremental tri-color
mark-sweep: every object is white (unvisited), gray (visited, children not yet scanned),
or black (fully scanned). Mark from roots, moving white→gray→black; sweep frees remaining
white. For a *single-threaded, stop-the-world* collector the tri-color invariant is
trivially maintained (no write barrier needed) — the three colors are just the worklist
state of a BFS/DFS. Lua's incremental variant flips between "two whites" to make sweeping
resumable, but you don't need that initially.
- https://poga.github.io/lua53-notes/gc.html
- https://deepwiki.com/lua/lua/5.1-garbage-collection

**Recommendation for mon_lang.** For a single-threaded language whose heap values are
references to structs/arrays, with roots = stack + globals, build a **precise, non-moving,
stop-the-world mark-sweep** collector — the design Go launched with, made precise like
TinyGo, structured as Lua's tri-color BFS.
- *Why precise over conservative:* you own the whole toolchain and already emit type info
  in `mtypes`/`tackygen`. Precise GC needs no false-positive heuristics, no "does this int
  look like a pointer" hazard, and no whole-stack conservative scan. The cost is you must
  emit **stack maps** (which slots/registers hold pointers at each safepoint) and **type
  layout maps** (which fields of each struct are pointers). This is more compiler work but
  a *far* simpler and more correct runtime, and it's the honest Go-like design.
- *Pragmatic staging:* ship **`leaking` (bump allocator, no collection) first** — it's ~30
  lines, lets self-hosting proceed (compilers are batch programs; leaking is often fine),
  and de-risks everything else. Then add STW precise mark-sweep as a second milestone.
- *If stack maps prove too hard initially:* fall back to **conservative stack scanning
  with precise heap scanning** (scan the stack conservatively for potential roots, but use
  exact type layouts to trace the heap) — Boehm-lite. This removes the hardest part (stack
  maps) while keeping heap tracing exact. Many hobby/teaching GCs do exactly this.
- *Design specifics:* object header with mark bit + type-id (index into a layout table);
  free-list or size-segregated allocator; roots = walk the Go-style linked frame list or a
  shadow stack; trigger collection on allocation when heap exceeds a growing threshold
  (e.g. 2× live-after-last-GC). Non-moving avoids pointer fixup — keep it that way.

---

## 3. Module / package systems in small compiled languages

**Go (directory = package, capital = export).** A package is a *directory*; all files in
it share one `package` declaration and one namespace (splitting a file across files in the
dir is free). Export is by **capitalization** of the identifier — `Foo` is exported, `foo`
is not. No export lists, no per-item annotations. A `go.mod` names the module (the import
path root). This makes "extract a helper into a new file" zero-ceremony and export status
visible at every use site.
- https://go.dev/doc/code
- https://www.alexedwards.net/blog/an-introduction-to-packages-imports-and-modules

**Zig (`@import` returns a struct, `pub` exports).** `@import("file.zig")` evaluates the
file and returns a *struct type* whose `pub` declarations are its fields; you bind it to a
const (`const std = @import("std");`) and access members. Export is explicit per-item via
`pub`. No index file, no directory magic — a file *is* a value. Callers in other files
need no edits when you add a `pub` decl.
- https://zig.guide/language-basics/imports/

**Single-pass / cross-module types.** Single-pass compilers (TCC-style) struggle with
cross-module type references because they resolve as they parse. The standard fixes: (a) a
**separate resolution pass** over all imported files before codegen (mon_lang already has
a distinct `validate` phase, so it is *not* strictly single-pass — this is an advantage),
or (b) **header/interface files** (C, OCaml `.mli`) that declare cross-module types
up-front, or (c) build an import DAG and topologically compile dependencies first,
exposing each module's exported symbol table to its dependents.

**Recommendation for mon_lang.** Keep the current `ашигла "file.mn"` + `тунх` model — it's
essentially Zig's (file-based, explicit-export) and is the right fit for a teaching
language and for self-hosting. Concrete guidance:
- **Stay file-based, not directory-based (for now).** Go's directory-as-package needs a
  module manifest and build system to resolve import paths; Zig's file-as-value needs
  none. A self-hosting compiler wants the *fewest* moving parts. `тунх`-explicit export is
  also more teachable than "capitalization has semantics."
- **Resolve modules as a DAG before codegen.** Parse each imported file, build the import
  graph, reject cycles (or handle them explicitly), topologically order, and run `validate`
  so every module's exported symbol/type table is available to importers. Your existing
  separate `validate` phase makes cross-module types a non-issue — do *not* regress into
  single-pass resolution.
- **Namespace exports to avoid collisions.** As programs grow (and the compiler imports
  many files), flat symbol merging breaks. Adopt Zig's binding: `ашигла` yields a named
  handle and members are accessed through it (`лексер.Токен`), rather than dumping every
  `тунх` symbol into the global scope. This is the single highest-value upgrade for
  scaling to a self-hosted multi-file compiler.
- **Defer a package *manager*** (fetch/versioning) indefinitely — it's orthogonal to
  self-hosting and adds large surface area.

---

## 4. Floating point lowering without libc

**Printing doubles — the algorithm ladder.** The shortest-round-trip problem
(print the fewest digits that parse back exactly) evolved: Dragon4 (Steele & White,
correct but needs bignums) → Grisu2/3 (Loitsch 2010, machine integers, but Grisu3 must
detect ~0.5% of cases it can't shorten and fall back to Dragon) → **Ryū** (Adams 2018,
~3× faster than Grisu3, uses 128-bit precomputed tables and fixed-precision integer
math only, always shortest + correctly rounded). Go is moving to an "unrounded scaling"
variant (single 64-bit multiply in the common case, table-based, no bignum) in ~1.27.
- https://github.com/ulfjack/ryu
- https://research.swtch.com/fp

**The simple correct fallback (rsc's key point).** You do **not** need shortest-digits to
be correct. **Fixed-width 17 significant digits** round-trips *any* IEEE-754 double
exactly. It uses the same table-based scaling infrastructure but skips the
shortest-boundary computation, avoids bignums, and is far simpler than Ryū. The only cost
is ugly output (`0.1` prints as `0.10000000000000001`). Traditional `%g` (repeated
division by 10) is even simpler but slower and needs care around rounding.
- https://research.swtch.com/fp

**Parsing doubles.** Modern answer is **Eisel–Lemire** (Lemire 2021, "Number Parsing at a
Gigabyte per Second"): ~80 lines of core code + a powers-of-10 table, one 128-bit multiply
with precomputed powers of 5, matches `strtod` bit-for-bit on ~99% of inputs and detects
the rest to fall back. Adopted by Rust, Go (1.16, ported by Nigel Tao into
`strconv.ParseFloat`), and Abseil `from_chars`. Simpler correct fallback: a careful
decimal-to-`double` via `double` accumulation is *not* always correctly rounded — avoid it
for a correct compiler; do the Lemire fast path with a slow-but-exact big-decimal fallback,
or just port Go's `strconv` logic since you're already in Go.
- https://nigeltao.github.io/blog/2020/eisel-lemire.html
- https://github.com/lemire/fast_float

**SSE2 — the minimum instruction set.** For scalar `double` (`тоо_бутархай`) on x86_64,
SSE2 is guaranteed present (part of the x86_64 baseline) and is what GCC/Clang/MSVC emit by
default. The minimal working set for a Sandler-style backend:
- `movsd` — load/store/move a 64-bit double (mem↔xmm, xmm↔xmm)
- `addsd` / `subsd` / `mulsd` / `divsd` — scalar arithmetic
- `comisd` (or `ucomisd`) — compare, sets EFLAGS (ZF/PF/CF) → drives `seta`/`setae`/etc.
  (note: unordered/NaN sets PF; float comparisons need unsigned-style condition codes)
- `cvttsd2si` — double→int with truncation (into a GPR); `cvtsi2sd` — int→double
- `cvtss2sd`/`cvtsd2ss` — only if you also support 32-bit `float`
- negation/abs via `xorpd`/`andpd` with a sign-mask constant in memory (there is no
  `negsd`); load double literals from a `.rodata`-style constant pool (you control the
  Mach-O writer, so emit a `__const` section and RIP-relative `movsd`).
- https://docs.oracle.com/cd/E26502_01/html/E28388/epmpv.html
- https://en.wikibooks.org/wiki/X86_Assembly/SSE

**Recommendation for mon_lang.**
1. **Codegen:** add a `double` type lowered to XMM registers with the SSE2 set above.
   Sandler's "Writing a C Compiler" Part II (doubles chapter) is the exact roadmap and
   matches your Tacky pipeline — follow it. Emit double literals via a constant pool in a
   `__const` Mach-O section, RIP-relative. Handle the NaN/unordered comparison codes
   carefully (this is the classic bug).
2. **Printing:** ship **fixed-width 17-digit** printing first (correct, round-trips, no
   bignum, small). Upgrade to shortest-round-trip (Ryū or rsc's unrounded-scaling) later
   for pretty output — it's a quality improvement, not correctness.
3. **Parsing:** port **Go's `strconv.ParseFloat`** (Eisel–Lemire fast path + exact
   fallback) — you're already in Go for the stage-0 compiler, and it's known-correct. When
   self-hosting, translate that routine into mon_lang.
4. Since you have no libc, all of the above must be **pure integer/SSE2 + your own tables**
   — which is exactly what Ryū/Lemire are (no libc math dependency). This is a *strength* of
   picking these algorithms: they need nothing from libm.

---

## 5. Register allocation at teaching-compiler scale

**What real small/fast compilers use.**
- **TCC** — no real allocator: single-pass, a value stack, and **three temporary
  registers**, allocated straightforwardly with no interference graph. Prioritizes
  compile speed over code quality; still produces working, reasonably fast code.
- **QBE** — a **linear-scan allocator with register hinting**; explicitly chosen as
  "simpler and faster than graph coloring." This is the sweet spot for a small
  optimizing backend.
- **Go** — an SSA-based allocator tuned for *compiler speed*; still the most expensive
  pass (~20% of the optimizer). Not graph coloring — Go deliberately avoids the quadratic
  cost.
- **Graph coloring** (Chaitin-Briggs; Sandler ch20 teaches this) — best code quality, but
  the coloring algorithm is roughly quadratic and is generally used in *batch* compilers
  where compile time is cheap. Linear scan is the JIT/fast-compiler default.
- https://c9x.me/compile/ (QBE)
- https://vnmakarov.github.io/2024/09/24/register-allocation-in-the-go-compiler.html
- http://web.cs.ucla.edu/~palsberg/course/cs132/linearscan.pdf (Poletto & Sarkar, linear scan)

**Tradeoffs / expected wins.** The big jump is **stack-slots → any register allocator at
all**: naive "every temporary lives in memory" spills constantly, and getting hot values
into registers is where most of the speedup lives. Between algorithms the gap is smaller:
linear scan "produces results often not dramatically worse than true graph coloring" while
being far faster to run (adding a linear-scan option cut OCaml's *compile* time ~25%; graph
coloring is "much slower … but near-optimal" *code*). Better *spill heuristics* (spill the
value that dies latest; register-pressure heuristics) buy ~15% on top of basic linear scan —
more than switching algorithms does.

**Recommendation for mon_lang.** Do **linear scan, not graph coloring**, despite Sandler
ch20 teaching coloring. Rationale:
- The dominant win is escaping stack slots; linear scan captures nearly all of it at a
  fraction of coloring's complexity and runtime. QBE — a respected small backend — made
  exactly this call.
- Linear scan is dramatically **simpler to implement and debug**, and simplicity compounds
  once *this allocator must itself be written in mon_lang* for self-hosting. A quadratic
  coloring pass is a liability in a self-hosted teaching compiler.
- Concrete plan: (1) compute live intervals over your Tacky/linear IR; (2) sort by start,
  scan, assign from a free pool of the callee/caller-saved GPRs (and XMMs for doubles once
  §4 lands), spilling the interval that ends latest when you run out; (3) keep the current
  stack-slot path as the guaranteed-correct fallback for anything the allocator punts on
  (aggregates, address-taken locals). (4) Add register hinting (QBE-style) to reduce
  move/copy instructions once the basics work.
- Read Sandler ch20 for the *interference/liveness* machinery and spilling discipline, but
  substitute the linear-scan assignment step for graph coloring. If you later want top code
  quality without full coloring, "linear scan on SSA form" (Wimmer & Franz) is the modern
  upgrade path.
- https://bernsteinbear.com/assets/img/wimmer-linear-scan-ssa.pdf

---

## Sources (primary)

- Go bootstrap: https://www.infoq.com/news/2015/01/golang-15-bootstrapped/
- Rust bootstrap: https://rustc-dev-guide.rust-lang.org/building/bootstrapping/what-bootstrapping-does.html
- Zig bootstrap: https://ziglang.org/news/goodbye-cpp/ , https://github.com/ziglang/zig/pull/13560
- Boehm GC: https://hboehm.info/gc/ , https://www.hboehm.info/gc/gcdescr.html
- TinyGo GC: https://aykevl.nl/2020/09/gc-tinygo/
- Lua GC: https://poga.github.io/lua53-notes/gc.html
- Go/Zig modules: https://go.dev/doc/code , https://zig.guide/language-basics/imports/
- Float print/parse: https://research.swtch.com/fp , https://github.com/ulfjack/ryu , https://nigeltao.github.io/blog/2020/eisel-lemire.html
- SSE2: https://docs.oracle.com/cd/E26502_01/html/E28388/epmpv.html
- Register allocation: https://c9x.me/compile/ , https://vnmakarov.github.io/2024/09/24/register-allocation-in-the-go-compiler.html
