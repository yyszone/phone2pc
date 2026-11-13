//go:build !windows

package sysutil

import (
	"fmt"
	"os"
)

func EnsureConsole() {}
func HideConsole()    {}

func ShowError(title, text string) {
	fmt.Fprintf(os.Stderr, "[%s] %s\n", title, text)
}