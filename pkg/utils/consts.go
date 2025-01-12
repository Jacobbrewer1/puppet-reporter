package utils

import (
	"reflect"
	"runtime"
	"strings"
)

const (
	appName          = "puppet_reporter"
	defaultVaultAddr = "http://vault-active.vault:8200"
)

func AppName(suffs ...string) string {
	if len(suffs) == 0 {
		return appName
	}
	return appName + "_" + strings.Join(suffs, "_")
}

// PackageName returns the package name of the given object
//
// Example:
// utils.PackageName(&service{}) // returns "api" and not "pkg/services/api"
func PackageName(of any) string {
	return runtime.FuncForPC(reflect.ValueOf(of).Pointer()).Name()
}
