package nfc

import "fmt"

// PrintInstructions 打印 NFC 标签写入说明
func PrintInstructions(url string) {
	fmt.Println("=================================================")
	fmt.Println("  NFC 碰一碰上传 —— 标签写入说明")
	fmt.Println("-------------------------------------------------")
	fmt.Println("  1. 在手机上安装任意 NFC 写入 App（如 NFC Tools）")
	fmt.Println("  2. 在 App 中选择写入「URL / 网址」")
	fmt.Println("  3. 把下面的地址写入你的 NFC 标签：")
	fmt.Println()
	fmt.Println("     " + url)
	fmt.Println()
	fmt.Println("  4. 之后手机碰一碰该标签，就会打开上传页面。")
	fmt.Println()
	fmt.Println("  提示：Android 碰一碰通常直接打开浏览器；")
	fmt.Println("        iPhone 碰一碰会先弹通知，需要手动点一下。")
	fmt.Println("=================================================")
	fmt.Println()
}