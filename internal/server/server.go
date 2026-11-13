package server

import (
	"fmt"
	"log"
	"net/http"

	"phone2pc/internal/config"
	"phone2pc/internal/netutil"
	"phone2pc/internal/qrcode"
	"phone2pc/internal/uploader"
)

// Server 封装 HTTP 服务与首页所需的 URL/二维码
type Server struct {
	url      string
	qrBase64 string
}

// New 计算访问 URL，并预生成二维码
func New() (*Server, error) {
	cfg := config.Get()
	ip := netutil.GetOutboundIP()
	url := "http://" + ip + ":" + cfg.Port

	qr, err := qrcode.GenerateBase64(url)
	if err != nil {
		return nil, fmt.Errorf("生成二维码失败: %w", err)
	}
	return &Server{url: url, qrBase64: qr}, nil
}

// Start 注册路由并开始监听
func (s *Server) Start() error {
	cfg := config.Get()
	addr := ":" + cfg.Port

	mux := http.NewServeMux()

	// 页面
	mux.HandleFunc("/", s.indexHandler)

	// 上传
	mux.HandleFunc("/upload", uploader.Handler)

	// 已上传文件静态访问： /files/xxx
	mux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(cfg.UploadDir))))

	// API
	mux.HandleFunc("/api/config/dir", s.apiSetDir)
	mux.HandleFunc("/api/config/port", s.apiSetPort)
	mux.HandleFunc("/api/files/clear", s.apiClear) // 精确匹配优先
	mux.HandleFunc("/api/files/", s.apiDelete)    // 前缀匹配
	mux.HandleFunc("/api/shutdown", s.apiShutdown)

	log.Printf("服务启动: %s", s.url)
	return http.ListenAndServe(addr, mux)
}

// URL 暴露给外部（用于终端打印等）
func (s *Server) URL() string { return s.url }