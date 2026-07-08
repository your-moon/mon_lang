#!/bin/sh
# Bootstrap driver: build mc (the mon_lang compiler written in mon_lang)
# with the Go-hosted compiler, then use mc to compile a program.
#
#   ./selfhost/build.sh <program.mn> <output>
# The program's entry function must be named гол.
set -e
here=$(cd "$(dirname "$0")/.." && pwd)
monc=${MONC:-$here/mon_lang}
"$monc" gen "$here/selfhost/mc.mn" -o /tmp/mc >/dev/null
/tmp/mc "$1"
cc -arch x86_64 out.s "$here/stdlib/cc/lib.c" -o "$2"
echo "built $2"
