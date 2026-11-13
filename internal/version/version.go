// Package version 保存通过 Makefile -ldflags 注入的构建信息
package version

var (
	Version    = "dev"           // 版本号
	GitCommit  = "unknown"       // git commit 短哈希（dirty 时会带 -dirty）
	GitSubject = "unknown"       // 最新 commit 的标题
	BuildTime  = "unknown"       // 构建时间（UTC）
)