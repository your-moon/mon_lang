/*
 * mon_lang - util
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package util

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
