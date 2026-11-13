package main

import (
	"fmt"
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
	"phone2pc/internal/version"
)

func main() {
	// 1. 加载配置（首次运行会生成 config.json）
	if err := config.Load(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	cfg := config.Get()

	// 2. 确保上传目录存在
	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatalf("创建上传目录 %s 失败: %v", cfg.UploadDir, err)
	}

	// 3. 组装 URL
	ip := netutil.GetOutboundIP()
	url := "http://" + ip + ":" + cfg.Port

	// 4. 启动信息
	fmt.Println("=================================================")
	fmt.Println("  手机传图到电脑")
	fmt.Println("  Version :", version.Version)
	fmt.Println("  Commit  :", version.GitCommit)
	fmt.Println("  Subject :", version.GitSubject)
	fmt.Println("  Build   :", version.BuildTime)
	fmt.Println("-------------------------------------------------")
	fmt.Println("  访问地址:", url)
	fmt.Println("  上传目录:", config.AbsUploadDir())
	fmt.Println("=================================================")
	qrcode.PrintTerminal(url)
	nfc.PrintInstructions(url)

	// 5. 创建服务
	srv, err := server.New()
	if err != nil {
		log.Fatalf("创建服务失败: %v", err)
	}

	// 6. 自动打开浏览器（异步、失败不影响使用）
	go func() {
		time.Sleep(500 * time.Millisecond)
		netutil.OpenBrowser(url)
	}()

	// 7. 优雅退出
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		fmt.Println("\n正在退出...")
		os.Exit(0)
	}()

	// 8. 启动 HTTP
	if err := srv.Start(); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}