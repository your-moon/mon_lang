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

import _ "embed"

//go:embed prelude.mn
var Prelude string
