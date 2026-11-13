# phone2pc

**手机扫码 / NFC 碰一碰，把图片传到电脑。手机端零安装。**

一个单二进制运行的局域网文件传输工具：电脑上跑一个 Go 服务，网页展示二维码；手机用系统相机扫码，或用 NFC 碰一碰标签，打开浏览器就能上传图片。文件统一落到电脑的 `uploadFile/` 目录。

## 功能

- 📱 **扫码上传** — 手机系统相机 / 扫一扫直接打开上传页，不装任何 App
- 🏷️ **NFC 碰一碰** — 把同一个 URL 写进 NFC 贴纸，碰一下就打开上传页
- 🖼️ **多文件批量上传** — 一次选择多张图片
- 📁 **在线改上传路径** — 网页上直接改，立即生效
- 🔌 **在线改端口** — 保存到 `config.json`，重启后生效
- 🗂️ **文件管理** — 列表、下载、单个删除、一键清空
- ⛔ **一键停服** — 网页按钮直接退出程序
- 🧾 **版本信息** — 网页底部和启动终端都显示 Version / Commit / Build
- 🚀 **单二进制** — CGO 关闭，跨平台编译，双击即用
- 💡 **轻量** — 无后台定时任务，无文件监听，首页仅做一次目录扫描

## 快速开始

```bash
git clone <your-repo> phone2pc
cd phone2pc
go mod tidy
go run .
```

启动后终端打印二维码 + NFC 写入说明，浏览器会自动打开控制台页面（二维码、上传、配置、文件管理）。

手机和电脑连同一个 Wi-Fi，用系统相机扫终端二维码，或手动打开终端打印的地址即可。

## 编译

```bash
make                # 测试 + 编译所有平台 + 打包 + SHA256（默认）
make version        # 查看当前自动计算的版本号
make linux-amd64    # 只编译 Linux amd64
make windows        # Windows amd64 + arm64
make macos          # macOS Intel + Apple Silicon
make tag            # 用当前计算的版本号打 git tag
make clean          # 清理 dist/
```

产物在 `dist/`，压缩包在 `dist/packages/`，校验值在 `dist/SHA256SUMS`。

版本号规则：

- 有 `VERSION=x.y.z` 环境变量 → 直接用
- HEAD 正好在 tag 上 → 用该 tag
- HEAD 在最近 tag 之后 → patch 自动 +1
- 完全无 tag → 从 `0.0.1` 起步

```bash
make VERSION=2.0.0     # 手动指定
make tag                # 发布时打 tag
```

## 使用

### 扫码

手机用系统相机或“扫一扫”扫描二维码，浏览器打开即可选图上传。

### NFC 碰一碰

NFC 贴纸本身就是一个存 URL 的 NDEF 标签。写入方式有两种：

1. **手机 App 写入（推荐）**：在手机装 NFC Tools（Android / iOS 均有），选择“写 URL”，把程序启动时终端打印的地址写进去。写入一次永久可用（只要电脑 IP 不变）。
2. **PN532 读卡器写入**：如果你有 PN532 模块，可以参考 `internal/nfc/nfc.go` 里的注释扩展。

注意事项：

- Android 碰一碰直接打开浏览器
- **iPhone 碰一碰会先弹通知，需要手动点击**，这是 iOS 的限制，不是程序问题
- 部分国产 ROM 默认关闭“NFC 网页跳转”，需在系统设置中打开

### 配置

程序首次运行会在当前目录生成 `config.json`：

```json
{
  "port": "8080",
  "uploadDir": "./uploadFile"
}
```

- `port`：服务监听端口，改完重启生效
- `uploadDir`：文件保存目录，相对路径基于程序运行目录，也可以用绝对路径（如 `D:\Receive`）

网页上也能改这两项，改完自动写入 `config.json`。

### 文件保存路径

默认是程序所在目录下的 `uploadFile/`。重名文件会自动加序号（`photo.jpg` → `photo_1.jpg`），不覆盖旧文件。文件名中的非法字符会被替换为 `_`，路径穿越会被拦截。

## HTTP 接口

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/` | 首页（二维码、配置、文件列表、版本号） |
| POST | `/upload` | 文件上传（`multipart/form-data`，字段名 `files`） |
| GET | `/files/<name>` | 下载已上传的文件 |
| POST | `/api/config/dir` | 修改上传路径，参数 `dir` |
| POST | `/api/config/port` | 修改端口，参数 `port` |
| DELETE | `/api/files/<name>` | 删除单个文件 |
| POST | `/api/files/clear` | 清空所有文件 |
| POST | `/api/shutdown` | 停止服务（`os.Exit(0)`） |

## 项目结构

```
phone2pc/
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── .gitignore
├── main.go                          # 入口：装配所有模块
└── internal/                        # Go 私有包，外部无法 import
    ├── version/version.go           # 版本变量（ldflags 注入）
    ├── config/config.go             # config.json 读写
    ├── netutil/network.go           # 局域网 IP、打开浏览器
    ├── qrcode/qr.go                 # 二维码（终端 ASCII + base64）
    ├── nfc/nfc.go                   # NFC 写入说明 + PN532 扩展点
    ├── uploader/
    │   ├── storage.go               # 文件列表 / 保存 / 删除 / 清空
    │   └── upload.go                # HTTP 上传处理器
    └── server/
        ├── server.go                # HTTP 服务与路由
        ├── handlers.go              # 首页 + API 处理器
        ├── templates.go             # embed 模板加载
        └── templates/
            └── index.html           # 首页（二维码、配置、文件列表、版本号）
```

运行时生成（已在 `.gitignore` 中忽略）：

```
config.json     # 配置
uploadFile/     # 收到的文件
dist/           # 编译产物
```

## 常见问题

**扫码后页面打不开？**
手机和电脑要在同一 Wi-Fi；Windows 首次运行会弹防火墙提示，需勾选“专用网络”；路由器若开启了 AP 隔离，同网段设备也访问不了。

**微信扫码上传失败？**
微信内置浏览器对 HTTP 表单上传有限制，改用系统相机扫码，或复制链接到 Chrome / Safari 打开。

**NFC 碰一碰没反应？**
iPhone 需要手动点通知；部分国产 ROM 需在设置中开启“NFC 网页跳转”；NFC 贴纸容量 ≥ 144 字节（NTAG213 即可）。

**`make` 报 `missing separator`？**
本 Makefile 用 `.RECIPEPREFIX := >` 代替 TAB，GNU Make 4.0+ 可直接用。如果 `make --version` 显示 3.x，把每个 `> ` 手动替换为 TAB。

**编译出来的程序在别的机器上打不开？**
默认 `CGO_ENABLED=0`，单文件静态编译，Linux / macOS / Windows 互不依赖。macOS 上如果是 Apple Silicon 下载的是 amd64 版本，会通过 Rosetta 运行，建议按平台选择。

**修改端口后不生效？**
端口写入 `config.json`，需重启程序。上传路径修改是即时生效的。

## License

MIT