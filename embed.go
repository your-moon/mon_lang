package main

import (
	_ "embed"

	"github.com/your-moon/mon_lang/linker"
	semanticanalysis "github.com/your-moon/mon_lang/semantic_analysis"
)

// The standard library is embedded in the compiler binary, so programs
// compile from any directory without extern declarations. The files live in
// stdlib/, which cannot hold a Go file because Go rejects .c files in a
// package directory without cgo.

//go:embed stdlib/prelude.mn
var embeddedPrelude string

//go:embed stdlib/lib.c
var embeddedLibC []byte

//go:embed stdlib/math.mn
var embeddedMath string

func init() {
	semanticanalysis.EmbeddedPrelude = embeddedPrelude
	semanticanalysis.EmbeddedStdlib["math.mn"] = embeddedMath
	linker.EmbeddedLibC = embeddedLibC
}
