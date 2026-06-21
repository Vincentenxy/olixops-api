package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

const sockPath = "/tmp/firecracker.sock"
const baseURL = "http://localhost"

// 基于 Unix Socket 创建 HTTP Client
func newUnixClient() *http.Client {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// 强制走 unix 域套接字
			return net.Dial("unix", sockPath)
		},
	}
	return &http.Client{Transport: transport}
}

func putAPI(path string, body any) error {
	cli := newUnixClient()
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, baseURL+path, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("api failed, status: %s", resp.Status)
	}
	return nil
}

func main() {
	// 替换成你本地内核、rootfs 绝对路径
	kernelPath := "/home/vincent/firecracker/assets/vmlinux.bin"
	rootfsPath := "/home/vincent/firecracker/assets/rootfs.ext4"

	// 配置内核
	bootSource := map[string]string{
		"kernel_image_path": kernelPath,
		"boot_args":         "console=ttyS0",
	}
	if err := putAPI("/boot-source", bootSource); err != nil {
		fmt.Printf("set boot source err: %v\n", err)
		return
	}
	fmt.Println("Set boot source success")

	// 配置 CPU、内存
	machineCfg := map[string]int{
		"vcpu_count":   1,
		"mem_size_mib": 512,
	}
	if err := putAPI("/machine-config", machineCfg); err != nil {
		fmt.Printf("set machine config err: %v\n", err)
		return
	}
	fmt.Println("Set machine config success")

	// 配置根磁盘
	drive := map[string]any{
		"drive_id":       "rootfs",
		"path_on_host":   rootfsPath,
		"is_root_device": true,
		"is_read_only":   false,
	}
	if err := putAPI("/drives/rootfs", drive); err != nil {
		fmt.Printf("set drive err: %v\n", err)
		return
	}
	fmt.Println("Set rootfs drive success")

	// 启动 MicroVM
	startAction := map[string]string{
		"action_type": "InstanceStart",
	}
	if err := putAPI("/actions", startAction); err != nil {
		fmt.Printf("start vm err: %v\n", err)
		return
	}
	fmt.Println("MicroVM started")
}
