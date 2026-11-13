//go:build windows

package sysutil

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	user32               = syscall.NewLazyDLL("user32.dll")
	procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")
	procShowWindow       = user32.NewProc("ShowWindow")
	procMessageBoxW      = user32.NewProc("MessageBoxW")
	procAttachConsole    = kernel32.NewProc("AttachConsole")
)

// EnsureConsole 如果编译为 GUI 程序并在终端中运行，尝试挂载到父进程的控制台
func EnsureConsole() {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd != 0 {
		return // 已经有控制台（比如普通编译），直接返回
	}

	// 尝试附加到父进程控制台 (ATTACH_PARENT_PROCESS = -1)
	const ATTACH_PARENT_PROCESS = ^uintptr(0)
	r, _, _ := procAttachConsole.Call(ATTACH_PARENT_PROCESS)
	if r == 0 {
		return // 附加失败（可能是双击运行），保持无控制台状态
	}

	// 重新绑定标准输入输出，否则 fmt.Println 没有反应
	out, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
	if err == nil {
		os.Stdout = out
		os.Stderr = out
	}
	in, err := os.OpenFile("CONIN$", os.O_RDONLY, 0)
	if err == nil {
		os.Stdin = in
	}
}

// HideConsole 隐藏当前的控制台黑窗口（降级处理）
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