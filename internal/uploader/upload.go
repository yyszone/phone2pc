package uploader

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"os"

	"phone2pc/internal/config"
)

const (
	maxUploadBytes = 500 << 20 // 500 MB
	memoryBufBytes = 32 << 20  // 32 MB
)

// Handler 处理 POST /upload
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	dir := config.Get().UploadDir
	if err := os.MkdirAll(dir, 0o755); err != nil {
		http.Error(w, "上传目录不可用: "+err.Error(), http.StatusInternalServerError)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(memoryBufBytes); err != nil {
		http.Error(w, "解析上传内容失败: "+err.Error(), http.StatusBadRequest)
		return
	}
	if r.MultipartForm == nil {
		http.Error(w, "请求不是 multipart/form-data", http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	fileHeaders := r.MultipartForm.File["files"]
	if len(fileHeaders) == 0 {
		http.Error(w, "没有收到文件", http.StatusBadRequest)
		return
	}

	var saved []string
	for _, fh := range fileHeaders {
		src, err := fh.Open()
		if err != nil {
			log.Printf("打开上传文件失败 %q: %v", fh.Filename, err)
			continue
		}
		name, n, err := Save(dir, fh.Filename, src)
		src.Close()
		if err != nil {
			log.Printf("保存 %s 失败: %v", fh.Filename, err)
			continue
		}
		log.Printf("已保存: %s/%s (%d 字节)", dir, name, n)
		saved = append(saved, name)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if len(saved) == 0 {
		_, _ = fmt.Fprint(w, resultPage("上传失败", nil))
		return
	}
	_, _ = fmt.Fprint(w, resultPage("上传成功", saved))
}

func resultPage(title string, names []string) string {
	s := `<!DOCTYPE html><html lang="zh-CN"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>` + html.EscapeString(title) + `</title>
<style>
body{font-family:system-ui,-apple-system,sans-serif;background:#f5f6f8;margin:0;padding:1.5rem;color:#222}
.card{max-width:520px;margin:auto;background:#fff;border-radius:12px;padding:1.5rem;box-shadow:0 2px 12px rgba(0,0,0,.08)}
h2{margin-top:0;color:#1677ff;font-size:1.2rem}
ul{padding-left:1.2rem;line-height:1.8;word-break:break-all}
a.btn{display:block;text-align:center;margin-top:1.2rem;padding:.9rem;background:#1677ff;color:#fff;text-decoration:none;border-radius:10px}
</style></head><body><div class="card">
<h2>` + html.EscapeString(title) + `</h2>`
	if len(names) > 0 {
		s += `<p>已保存到电脑的 <code>uploadFile</code> 目录：</p><ul>`
		for _, n := range names {
			s += `<li>` + html.EscapeString(n) + `</li>`
		}
		s += `</ul>`
	}
	s += `<a class="btn" href="/">继续上传</a></div></body></html>`
	return s
}