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
