package server

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"phone2pc/internal/config"
	"phone2pc/internal/uploader"
	"phone2pc/internal/version"
)

// pageData 首页模板数据结构
type pageData struct {
	URL          string
	QRBase64     string
	UploadDir    string
	AbsUploadDir string
	Port         string
	Files        []uploader.FileInfo
	Version      string
	GitCommit    string
	GitSubject   string
	BuildTime    string
}

// indexHandler 首页
func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	cfg := config.Get()
	files, _ := uploader.List(cfg.UploadDir)

	data := pageData{
		URL:          s.url,
		QRBase64:     s.qrBase64,
		UploadDir:    cfg.UploadDir,
		AbsUploadDir: config.AbsUploadDir(),
		Port:         cfg.Port,
		Files:        files,
		Version:      version.Version,
		GitCommit:    version.GitCommit,
		GitSubject:   version.GitSubject,
		BuildTime:    version.BuildTime,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := indexTmpl.Execute(w, data); err != nil {
		http.Error(w, "模板渲染失败: "+err.Error(), http.StatusInternalServerError)
	}
}

// ---------- API ----------

func writeJSON(w http.ResponseWriter, ok bool, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": ok, "msg": msg})
}

// apiSetDir POST /api/config/dir
func (s *Server) apiSetDir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, false, "Method Not Allowed")
		return
	}
	_ = r.ParseForm()
	dir := strings.TrimSpace(r.FormValue("dir"))
	if err := config.SetUploadDir(dir); err != nil {
		writeJSON(w, false, err.Error())
		return
	}
	writeJSON(w, true, "上传路径已更新: "+dir)
}

// apiSetPort POST /api/config/port
func (s *Server) apiSetPort(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, false, "Method Not Allowed")
		return
	}
	_ = r.ParseForm()
	port := strings.TrimSpace(r.FormValue("port"))
	if err := config.SetPort(port); err != nil {
		writeJSON(w, false, err.Error())
		return
	}
	writeJSON(w, true, "端口已保存为 "+port+"，重启程序后生效")
}

// apiDelete DELETE /api/files/<name>
func (s *Server) apiDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSON(w, false, "Method Not Allowed")
		return
	}
	raw := strings.TrimPrefix(r.URL.Path, "/api/files/")
	name, err := url.PathUnescape(raw)
	if err != nil || name == "" {
		writeJSON(w, false, "非法文件名")
		return
	}
	name = filepath.Base(name)
	if name == "." || name == ".." {
		writeJSON(w, false, "非法文件名")
		return
	}
	if err := uploader.Delete(config.Get().UploadDir, name); err != nil {
		writeJSON(w, false, "删除失败: "+err.Error())
		return
	}
	writeJSON(w, true, "已删除 "+name)
}

// apiClear POST /api/files/clear
func (s *Server) apiClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, false, "Method Not Allowed")
		return
	}
	if err := uploader.Clear(config.Get().UploadDir); err != nil {
		writeJSON(w, false, "清空失败: "+err.Error())
		return
	}
	writeJSON(w, true, "已清空所有文件")
}

// apiShutdown POST /api/shutdown
func (s *Server) apiShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, false, "Method Not Allowed")
		return
	}
	writeJSON(w, true, "服务即将停止")
	go func() {
		time.Sleep(500 * time.Millisecond)
		os.Exit(0)
	}()
}