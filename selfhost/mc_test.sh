#!/bin/sh
# self-hosting smoke test: mc compiles programs and they run correctly
set -e
here=$(cd "$(dirname "$0")/.." && pwd)
"${MONC:-$here/mon_lang}" gen "$here/selfhost/mc.mn" -o /tmp/mc >/dev/null
tmp=$(mktemp -d)
cat > "$tmp/t.mn" <<'PROG'
функц факт(н тоо64) -> тоо64 {
    хэрэв н <= 1 бол { буц 1; }
    буц н * факт(н - 1);
}
функц үндсэн() -> тоо64 {
    зарла и: тоо64 = 0;
    зарла с: тоо64 = 0;
    давтах и < 6 бол { с = с + и; и = и + 1; }
    хэвлэ(факт(5)); мөр_хэвлэх(" "); хэвлэ(с); мөр_хэвлэх("\n");
    буц 0;
}
PROG
cd "$tmp"
/tmp/mc "$tmp/t.mn"
cc -arch x86_64 out.s "$here/stdlib/cc/lib.c" -o prog
got=$(arch -x86_64 ./prog)
[ "$got" = "120 15" ] && echo "PASS: mc compiled correctly ($got)" || { echo "FAIL: got '$got'"; exit 1; }
