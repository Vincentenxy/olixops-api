package agent

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"olixops/internal/platform/logger"
	"os"

	"go.uber.org/zap"
)

// Firecracker UDS 路径
const sockPath = "/tmp/firecracker.socket"

func main() {

	log := logger.L()

	log.Info("Starting agent")

	// 1. 自定义 Transport：使用 Unix Domain Socket 拨号
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				// 强制走 unix socket，忽略原 addr
				return net.Dial("unix", sockPath)
			},
		},
	}

	// 2. 请求 Firecracker API，http://localhost 只是占位符
	// 示例1：获取虚拟机实例信息
	resp, err := client.Get("http://localhost/v1/instance")
	if err != nil {
		log.Error("Error getting instance", zap.Error(err))
		os.Exit(1)
	}
	defer resp.Body.Close()

	// 3. 读取并打印响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		os.Exit(1)
	}

	log.Info("状态码:  ",
		zap.Int("statusCode", resp.StatusCode),
	)
	log.Info("响应内容：", zap.String("body", string(body)))

}
