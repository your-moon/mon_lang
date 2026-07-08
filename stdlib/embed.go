/*
 * mon_lang - embedded standard library prelude
 *
 * Copyright (c) 2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

// The prelude (builtin function declarations) is compiled into the compiler
// binary, so mon_lang works from any directory as a single self-contained
// executable. On-disk stdlib files remain only for the legacy --cc path.
package stdlib

import (
	"embed"
)

//go:embed prelude.mn
var Prelude string

//go:embed pkg
var pkgFS embed.FS

// StdPackage returns the source of a standard-library package by bare name
// (e.g. "файл"), and whether it exists. Bare-name imports resolve here so
// standard packages work from any directory, like Go's import paths.
func StdPackage(name string) (string, bool) {
	data, err := pkgFS.ReadFile("pkg/" + name + ".mn")
	if err != nil {
		return "", false
	}
	return string(data), true
}
