/*
 * mon_lang - util/roundingutil
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package roundingutil

func RoundAwayFromZero(n, x int) int {
	if x%n == 0 {
		return x
	} else if x < 0 {
		return x - n - (x % n)
	} else {
		return x + n - (x % n)
	}
}
