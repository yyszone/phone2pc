package dialog

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// SelectDirectory 调用系统原生文件夹选择框
func SelectDirectory() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return runWindowsDialog("powershell")
	case "darwin":
		return runMacDialog()
	default:
		// 判断是否在 WSL 中
		if isWSL() {
			// WSL 下调用 Windows 的对话框
			path, err := runWindowsDialog("powershell.exe")
			if err != nil {
				return "", err
			}
			return winToWSLPath(path), nil
		}
		return runLinuxDialog()
	}
}

func runWindowsDialog(psCmd string) (string, error) {
	// 使用 PowerShell 调用 .NET 的 FolderBrowserDialog
	script := `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.FolderBrowserDialog; $f.Description = "请选择文件保存路径"; if($f.ShowDialog() -eq "OK") { Write-Output $f.SelectedPath }`
	cmd := exec.Command(psCmd, "-NoProfile", "-Command", script)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	path := strings.TrimSpace(out.String())
	if path == "" {
		return "", fmt.Errorf("未选择目录或被取消")
	}
	return path, nil
}

func runMacDialog() (string, error) {
	script := `POSIX path of (choose folder with prompt "请选择文件保存路径")`
	cmd := exec.Command("osascript", "-e", script)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	path := strings.TrimSpace(out.String())
	if path == "" {
		return "", fmt.Errorf("未选择目录或被取消")
	}
	return path, nil
}

func runLinuxDialog() (string, error) {
	// 依赖 zenity，如果没有则报错
	if _, err := exec.LookPath("zenity"); err != nil {
		return "", fmt.Errorf("系统缺少 zenity，无法打开选择框，请手动输入路径")
	}
	cmd := exec.Command("zenity", "--file-selection", "--directory", "--title=请选择文件保存路径")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	path := strings.TrimSpace(out.String())
	if path == "" {
		return "", fmt.Errorf("未选择目录或被取消")
	}
	return path, nil
}

func isWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), "microsoft")
}

// winToWSLPath 将 Windows 路径 (D:\foo) 转换成 WSL 路径 (/mnt/d/foo)
func winToWSLPath(winPath string) string {
	if len(winPath) > 1 && winPath[1] == ':' {
		drive := strings.ToLower(string(winPath[0]))
		rest := strings.ReplaceAll(winPath[2:], "\\", "/")
		return "/mnt/" + drive + rest
	}
	return winPath
}