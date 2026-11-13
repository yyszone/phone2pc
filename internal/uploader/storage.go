package uploader

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FileInfo 用于前端列表展示
type FileInfo struct {
	Name       string
	Size       int64
	SizeStr    string
	ModUnix    int64
	ModTimeStr string
}

// List 返回 dir 下所有文件的元信息，按修改时间倒序
func List(dir string) ([]FileInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	out := make([]FileInfo, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, FileInfo{
			Name:       e.Name(),
			Size:       info.Size(),
			SizeStr:    humanSize(info.Size()),
			ModUnix:    info.ModTime().Unix(),
			ModTimeStr: info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ModUnix > out[j].ModUnix })
	return out, nil
}

// Save 把 src 写入 dir 下，返回最终文件名与写入字节数
func Save(dir, rawName string, src io.Reader) (string, int64, error) {
	name := uniqueName(dir, rawName)
	path := filepath.Join(dir, name)

	dst, err := os.Create(path)
	if err != nil {
		return "", 0, err
	}
	n, copyErr := io.Copy(dst, src)
	dst.Close()
	if copyErr != nil {
		_ = os.Remove(path)
		return "", 0, copyErr
	}
	return name, n, nil
}

// Delete 删除 dir 下指定文件（防路径穿越）
func Delete(dir, name string) error {
	return os.Remove(filepath.Join(dir, filepath.Base(name)))
}

// Clear 删除 dir 下所有文件（不删子目录）
func Clear(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
	return nil
}

// uniqueName 过滤非法字符，重名时自动加序号，不覆盖
func uniqueName(dir, rawName string) string {
	name := strings.ReplaceAll(rawName, "\\", "/")
	name = filepath.Base(name)
	name = strings.TrimSpace(name)
	name = strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', 0:
			return '_'
		}
		return r
	}, name)
	if name == "" || name == "." || name == ".." {
		name = "unnamed"
	}

	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	candidate := name
	for i := 1; ; i++ {
		if _, err := os.Stat(filepath.Join(dir, candidate)); os.IsNotExist(err) {
			return candidate
		}
		candidate = fmt.Sprintf("%s_%d%s", base, i, ext)
	}
}

// humanSize 人类可读大小
func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

// 保留 time 导入用于未来扩展，避免 unused 报错
var _ = time.Now