package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"phone2pc/internal/config"
	"phone2pc/internal/netutil"
	"phone2pc/internal/nfc"
	"phone2pc/internal/qrcode"
	"phone2pc/internal/server"
	"phone2pc/internal/sysutil"
	"phone2pc/internal/version"
)

func main() {
	// 1. 解析参数
	showConsole := flag.Bool("console", false, "是否保留控制台黑窗口（默认隐藏）")
	flag.Parse()

	// 2. 如果要求显示控制台，尝试接管父窗口（针对 -H windowsgui 编译）
	if *showConsole {
		sysutil.EnsureConsole()
	}

	// 3. 初始化日志（终端 + 文件）
	initLog()

	// 4. 如果不要求显示控制台，隐藏窗口（针对普通控制台编译双击闪窗）
	if !*showConsole {
		sysutil.HideConsole()
	}

	// 5. 加载配置
	if err := config.Load(); err != nil {
		fatal("加载配置失败", err)
	}
	cfg := config.Get()

	// 6. 确保目录存在
	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		fatal(fmt.Sprintf("创建上传目录 %s 失败", cfg.UploadDir), err)
	}

	// 7. 计算 URL
	ip := netutil.GetOutboundIP()
	url := "http://" + ip + ":" + cfg.Port

	// 8. 启动信息
	fmt.Println("=================================================")
	fmt.Println("  手机传图到电脑")
	fmt.Println("  Version :", version.Version)
	fmt.Println("  Commit  :", version.GitCommit)
	fmt.Println("-------------------------------------------------")
	fmt.Println("  访问地址:", url)
	fmt.Println("  上传目录:", config.AbsUploadDir())
	fmt.Println("=================================================")
	qrcode.PrintTerminal(url)
	nfc.PrintInstructions(url)

	// 9. 创建服务
	srv, err := server.New()
	if err != nil {
		fatal("创建服务失败", err)
	}

	// 10. 自动打开浏览器
	go func() {
		time.Sleep(500 * time.Millisecond)
		netutil.OpenBrowser(url)
	}()

	// 11. 优雅退出
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		fmt.Println("\n正在退出...")
		os.Exit(0)
	}()

	// 12. 启动 HTTP 服务
	if err := srv.Start(); err != nil {
		fatal("服务启动失败", err)
	}
}

func initLog() {
	writers := []io.Writer{os.Stderr}
	f, err := os.OpenFile("phone2pc.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err == nil {
		writers = append(writers, f)
	}
	log.SetOutput(io.MultiWriter(writers...))
}

func fatal(prefix string, err error) {
	msg := prefix
	if err != nil {
		msg += ": " + err.Error()
	}
	log.Println(msg)
	sysutil.ShowError("phone2pc 错误", msg)
	os.Exit(1)
}