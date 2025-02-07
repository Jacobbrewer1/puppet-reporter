//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"

	_ "github.com/go-sql-driver/mysql"
	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
)

// Default target to run when none is specified
// If not set, running mage will list available targets
var Default = Build

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	fmt.Println("Building...")
	cmd := exec.Command("go", "build", "-o", "MyApp", ".")
	return cmd.Run()
}

// A custom install step if you need your bin someplace other than go/bin
func Install() error {
	mg.Deps(Build)
	fmt.Println("Installing...")
	return os.Rename("./MyApp", "/usr/bin/MyApp")
}

// Clean up after yourself
func Clean() error {
	fmt.Println("Cleaning...")
	os.RemoveAll("bin")
	return nil
}
