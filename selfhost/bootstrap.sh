#!/bin/sh
# Full self-hosting bootstrap verification for mon_lang.
#   stage1: mc (built by the Go-hosted compiler) compiles mc.mn
#   stage2: assemble stage1 into mc2 (mc compiled by mc)
#   stage3: mc2 compiles mc.mn
# Self-hosting holds when stage1 == stage3 (a reproducing fixed point).
set -e
here=$(cd "$(dirname "$0")/.." && pwd)
monc=${MONC:-$here/mon_lang}
# Build the Go-hosted compiler fresh so we never bootstrap from a stale binary
# (an old mon_lang missing e.g. [] array syntax fails at stage 0).
if [ -z "$MONC" ]; then
    (cd "$here" && go build -o "$monc" .)
fi
lib="$here/stdlib/cc/lib.c"
mc="$here/selfhost/mc.mn"
tmp=$(mktemp -d); cd "$tmp"

"$monc" gen "$mc" -o "$tmp/mc1" >/dev/null      # stage 0: Go-hosted → mc1
"$tmp/mc1" "$mc"; cp out.s stage1.s                 # stage 1: mc1 compiles mc.mn
cc -arch x86_64 stage1.s "$lib" -o mc2         # stage 2: assemble → mc2
arch -x86_64 ./mc2 "$mc"; cp out.s stage3.s    # stage 3: mc2 compiles mc.mn

if diff -q stage1.s stage3.s >/dev/null; then
    echo "SELF-HOSTING OK: stage1 == stage3 ($(wc -l < stage1.s | tr -d ' ') lines, byte-identical)"
else
    echo "FAIL: stage1 != stage3"; exit 1
fi
