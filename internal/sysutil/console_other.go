//go:build !windows

package sysutil

import (
	"fmt"
	"os"
)

// HideConsole 非 Windows 平台不支持隐藏黑框，空实现
func HideConsole() {}

// ShowError 非 Windows 平台直接打印到终端
func ShowError(title, text string) {
	fmt.Fprintf(os.Stderr, "[%s] %s\n", title, text)
}