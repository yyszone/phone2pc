//go:build windows

package sysutil

import (
	"syscall"
	"unsafe"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	user32               = syscall.NewLazyDLL("user32.dll")
	procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")
	procShowWindow       = user32.NewProc("ShowWindow")
	procMessageBoxW      = user32.NewProc("MessageBoxW")
)

// HideConsole 隐藏当前的控制台黑窗口
func HideConsole() {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd != 0 {
		const SW_HIDE = 0
		procShowWindow.Call(hwnd, SW_HIDE)
	}
}

// ShowError 弹出一个 Windows 原生错误提示对话框
func ShowError(title, text string) {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	textPtr, _ := syscall.UTF16PtrFromString(text)
	const MB_ICONERROR = 0x00000010
	procMessageBoxW.Call(0, uintptr(unsafe.Pointer(textPtr)), uintptr(unsafe.Pointer(titlePtr)), MB_ICONERROR)
}