// cmd/server 是 olixops 的 HTTP 服务入口。
//
// 该文件保持极薄:只做参数解析与退出码处理,
// 真正的应用装配、生命周期、信号监听放在 internal/app。
package main

import (
	"fmt"
	"os"

	"olixops/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}
