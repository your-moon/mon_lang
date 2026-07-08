/*
 * mon_lang - util
 *
 * Copyright (c) 2024-2026 Munkherdene
 * SPDX-License-Identifier: MIT (see LICENSE)
 */

package util

import "runtime"

type OsType string

const (
	Linux   OsType = "linux"
	Darwin  OsType = "darwin"
	Windows OsType = "windows"
)

func GetOsType() OsType {
	return OsType(runtime.GOOS)
}
