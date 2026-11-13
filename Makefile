# phone2pc Makefile
# 需要 GNU Make 4.0+；用 .RECIPEPREFIX 避免 TAB 问题
# 版本号规则：git tag 自动递增，VERSION=x.y.z 可手动覆盖

.RECIPEPREFIX := >
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
.ONESHELL:
.DEFAULT_GOAL := package

GO       ?= go
DIST_DIR ?= dist
PKG      := phone2pc

# ---------- 版本计算 ----------
# 1) 有 VERSION 环境变量，直接用
# 2) HEAD 正好在某个 tag 上（git describe --exact-match），用该 tag
# 3) 否则取最近一个 tag，patch +1
# 4) 完全没有 tag，从 0.0.1 开始
VERSION ?= $(shell \
  if ! command -v git >/dev/null 2>&1 || ! git rev-parse --git-dir >/dev/null 2>&1; then \
    echo "0.0.1"; exit 0; \
  fi; \
  exact=$$(git describe --tags --exact-match 2>/dev/null || true); \
  if [ -n "$$exact" ]; then \
    echo "$$exact" | sed 's/^v//'; exit 0; \
  fi; \
  last=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
  echo "$$last" | sed 's/^v//' \
    | awk -F. '{if (NF>=3) print $$1"."$$2"."$$3+1; \
               else if (NF==2) print $$1"."$$2".1"; \
               else print $$0".0.1"}')

# ---------- Git 元信息 ----------
GIT_COMMIT  := $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo "source-archive")
GIT_SUBJECT := $(shell git log -1 --pretty=%s 2>/dev/null | tr ' ' '-' | tr -cd '[:alnum:]_.-' || echo "source-build")
GIT_DIRTY   := $(shell if [ -n "$$(git status --porcelain 2>/dev/null)" ]; then echo "-dirty"; fi)
GIT_COMMIT  := $(GIT_COMMIT)$(GIT_DIRTY)
ifeq ($(strip $(GIT_SUBJECT)),)
  GIT_SUBJECT := source-build
endif

BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
  -X $(PKG)/internal/version.Version=$(VERSION) \
  -X $(PKG)/internal/version.GitCommit=$(GIT_COMMIT) \
  -X $(PKG)/internal/version.GitSubject=$(GIT_SUBJECT) \
  -X $(PKG)/internal/version.BuildTime=$(BUILD_TIME)

.PHONY: all package binaries test version windows linux macos clean tag help

all: package

binaries: windows linux macos

windows: windows-amd64 windows-arm64
linux:   linux-amd64 linux-arm64 linux-armv7
macos:   macos-amd64 macos-arm64

# 每个 target 输出一个独立可执行文件到 $(DIST_DIR)
define GO_BUILD
.PHONY: $(1)
$(1):
> @mkdir -p "$(DIST_DIR)"
> echo "=====> Building $(4) (v$(VERSION))"
> CGO_ENABLED=0 GOOS=$(2) GOARCH=$(3) $(5) $(GO) build -trimpath -ldflags='$(LDFLAGS) $(6)' -o "$(DIST_DIR)/$(4)" .
endef


# Windows 追加 -H windowsgui
$(eval $(call GO_BUILD,windows-amd64,windows,amd64,phone2pc-windows-amd64.exe,,-H windowsgui))
$(eval $(call GO_BUILD,windows-arm64,windows,arm64,phone2pc-windows-arm64.exe,,-H windowsgui))

# 其他平台保持为空
$(eval $(call GO_BUILD,linux-amd64,linux,amd64,phone2pc-linux-amd64,))
$(eval $(call GO_BUILD,linux-arm64,linux,arm64,phone2pc-linux-arm64,))
$(eval $(call GO_BUILD,linux-armv7,linux,arm,phone2pc-linux-armv7,GOARM=7))
$(eval $(call GO_BUILD,macos-amd64,darwin,amd64,phone2pc-macos-amd64,))
$(eval $(call GO_BUILD,macos-arm64,darwin,arm64,phone2pc-macos-arm64,))

test:
> @echo "=====> Running tests"
> $(GO) test ./...

version:
> @echo "Version:  $(VERSION)"
> @echo "Commit:   $(GIT_COMMIT)"
> @echo "Subject:  $(GIT_SUBJECT)"
> @echo "Build:    $(BUILD_TIME)"

package: test binaries
> @mkdir -p "$(DIST_DIR)/packages"
> echo "=====> Creating release archives"
> if command -v zip >/dev/null 2>&1; then
>   (cd "$(DIST_DIR)" && zip -q -j "packages/phone2pc-windows-amd64.zip" "phone2pc-windows-amd64.exe")
>   (cd "$(DIST_DIR)" && zip -q -j "packages/phone2pc-windows-arm64.zip" "phone2pc-windows-arm64.exe")
> else
>   tar -czf "$(DIST_DIR)/packages/phone2pc-windows-amd64.tar.gz" -C "$(DIST_DIR)" "phone2pc-windows-amd64.exe"
>   tar -czf "$(DIST_DIR)/packages/phone2pc-windows-arm64.tar.gz" -C "$(DIST_DIR)" "phone2pc-windows-arm64.exe"
> fi
> for name in phone2pc-linux-amd64 phone2pc-linux-arm64 phone2pc-linux-armv7 phone2pc-macos-amd64 phone2pc-macos-arm64; do
>   tar -czf "$(DIST_DIR)/packages/$${name}.tar.gz" -C "$(DIST_DIR)" "$${name}"
> done
> if command -v sha256sum >/dev/null 2>&1; then
>   (cd "$(DIST_DIR)" && find . -maxdepth 2 -type f ! -name SHA256SUMS -print0 | sort -z | xargs -0 sha256sum) > "$(DIST_DIR)/SHA256SUMS"
> else
>   (cd "$(DIST_DIR)" && find . -maxdepth 2 -type f ! -name SHA256SUMS -print0 | sort -z | xargs -0 shasum -a 256) > "$(DIST_DIR)/SHA256SUMS"
> fi
> echo
> echo "=====> Build complete"
> echo "Version:  $(VERSION)"
> echo "Commit:   $(GIT_COMMIT)"
> echo "Subject:  $(GIT_SUBJECT)"
> echo "Build:    $(BUILD_TIME)"
> ls -lh "$(DIST_DIR)" "$(DIST_DIR)/packages"

tag:
> @if [ "$$(git status --porcelain)" != "" ]; then
>   echo "工作区有未提交改动，先提交再打 tag"
>   exit 1
> fi
> git tag -a "v$(VERSION)" -m "Release v$(VERSION)"
> echo "已创建 tag v$(VERSION)，执行 git push --tags 推送到远程"

clean:
> @echo "=====> Cleaning"
> rm -rf "$(DIST_DIR)"

help:
> @echo "phone2pc 构建目标："
> @echo "  make                测试 + 编译所有平台 + 打包（默认）"
> @echo "  make package        同 make"
> @echo "  make test           运行 go test ./..."
> @echo "  make binaries       只编译所有平台"
> @echo "  make windows        编译 Windows amd64 + arm64"
> @echo "  make linux          编译 Linux amd64 + arm64 + armv7"
> @echo "  make macos          编译 macOS Intel + Apple Silicon"
> @echo "  make version        显示当前计算的版本号"
> @echo "  make tag            为当前版本打 git tag"
> @echo "  make clean          清理 dist/"
> @echo "  make VERSION=1.2.0  手动指定版本号"
> @echo "  make DIST_DIR=out   修改输出目录"