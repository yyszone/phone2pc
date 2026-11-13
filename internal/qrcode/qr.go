package qrcode

import (
	"encoding/base64"
	"os"

	"github.com/mdp/qrterminal/v3"
	skip2 "github.com/skip2/go-qrcode"
)

// PrintTerminal 在终端打印 ASCII 二维码
func PrintTerminal(url string) {
	qrterminal.GenerateHalfBlock(url, qrterminal.L, os.Stdout)
}

// GenerateBase64 生成二维码 PNG 的 base64 编码，供网页内嵌
func GenerateBase64(url string) (string, error) {
	pngBytes, err := skip2.Encode(url, skip2.Medium, 256)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(pngBytes), nil
}