package netutil

import (
	"log"
	"net"
	"os/exec"
	"runtime"
)

// GetOutboundIP 通过一次 UDP "连接" 获取局域网出口 IP
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

// OpenBrowser 跨平台打开默认浏览器
func OpenBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("自动打开浏览器失败（可手动访问 %s）: %v", url, err)
	}
}