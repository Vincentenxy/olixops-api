package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
)

const (
	rootfsPath   = "/home/vincent/firecracker/rootfs.ext4"
	globalKernel = "/home/vincent/firecracker/vmlinux"
	vmRootDir    = "/var/lib/fc-vms"
	listenAddr   = "0.0.0.0:8080"
)

type VMStatus string

const (
	StatusIdle    VMStatus = "idle"
	StatusRunning VMStatus = "running"
	StatusStopped VMStatus = "stopped"
)

type VMInstance struct {
	VMID     string   `json:"vm_id"`
	SockPath string   `json:"sock_path"`
	PID      int      `json:"pid"`
	Status   VMStatus `json:"status"`
}

var (
	instanceMap = make(map[string]*VMInstance)
	mu          sync.RWMutex
)

func initDir() error {
	return os.MkdirAll(vmRootDir, 0755)
}

func getVMPaths(vmID string) (dir, sockPath string) {
	dir = filepath.Join(vmRootDir, vmID)
	sockPath = filepath.Join(dir, "fc.sock")
	return dir, sockPath
}

func fcRequest(sockPath, path string, body any) error {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", sockPath)
		},
	}
	client := &http.Client{Transport: transport}

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, "http://localhost"+path, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("api code: %d", resp.StatusCode)
	}
	return nil
}

// 代码拉起 Firecracker 进程（核心：创建多实例）
func startFirecrackerProcess(sockPath string) (int, error) {
	_ = os.Remove(sockPath)
	cmd := exec.Command("firecracker", "--api-sock", sockPath)
	// 后台运行，脱离终端
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	return cmd.Process.Pid, nil
}

// 杀死进程（销毁实例用）
func killProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(syscall.SIGKILL)
}

func createVMHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	vmID := r.URL.Query().Get("vm_id")
	if vmID == "" {
		http.Error(w, "missing vm_id", 400)
		return
	}

	mu.RLock()
	_, exists := instanceMap[vmID]
	mu.RUnlock()
	if exists {
		http.Error(w, "vm "+vmID+" exists", 409)
		return
	}

	vmDir, sockPath := getVMPaths(vmID)
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		http.Error(w, "create dir failed: "+err.Error(), 500)
		return
	}

	// 1. 代码启动 Firecracker 进程
	pid, err := startFirecrackerProcess(sockPath)
	if err != nil {
		http.Error(w, "start firecracker failed: "+err.Error(), 500)
		return
	}

	// 2. 配置内核
	boot := map[string]string{
		"kernel_image_path": globalKernel,
		"boot_args":         "console=ttyS0 reboot=k panic=1 pci=off",
	}
	if err := fcRequest(sockPath, "/boot-source", boot); err != nil {
		_ = killProcess(pid)
		http.Error(w, "config boot failed: "+err.Error(), 500)
		return
	}

	// 3. 配置硬件
	machine := map[string]int{
		"vcpu_count":   1,
		"mem_size_mib": 512,
	}
	if err := fcRequest(sockPath, "/machine-config", machine); err != nil {
		_ = killProcess(pid)
		http.Error(w, "config machine failed: "+err.Error(), 500)
		return
	}

	// 4. 配置根盘
	drive := map[string]any{
		"drive_id":       "rootfs",
		"path_on_host":   rootfsPath,
		"is_root_device": true,
		"is_read_only":   false,
	}
	if err := fcRequest(sockPath, "/drives/rootfs", drive); err != nil {
		_ = killProcess(pid)
		http.Error(w, "config drive failed: "+err.Error(), 500)
		return
	}

	// 5. 启动 VM
	action := map[string]string{"action_type": "InstanceStart"}
	if err := fcRequest(sockPath, "/actions", action); err != nil {
		_ = killProcess(pid)
		http.Error(w, "start vm failed: "+err.Error(), 500)
		return
	}

	// 6. 记录实例
	mu.Lock()
	instanceMap[vmID] = &VMInstance{
		VMID:     vmID,
		SockPath: sockPath,
		PID:      pid,
		Status:   StatusRunning,
	}
	mu.Unlock()

	w.WriteHeader(200)
	_, _ = fmt.Fprintf(w, "VM %s ok, pid: %d", vmID, pid)
}

func stopVMHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	vmID := r.URL.Query().Get("vm_id")
	if vmID == "" {
		http.Error(w, "missing vm_id", 400)
		return
	}

	mu.RLock()
	inst, exists := instanceMap[vmID]
	mu.RUnlock()
	if !exists {
		http.Error(w, "vm not found", 404)
		return
	}

	action := map[string]string{"action_type": "SendCtrlAltDel"}
	if err := fcRequest(inst.SockPath, "/actions", action); err != nil {
		http.Error(w, "stop failed: "+err.Error(), 500)
		return
	}

	mu.Lock()
	inst.Status = StatusStopped
	mu.Unlock()
	w.WriteHeader(200)
	_, _ = fmt.Fprintf(w, "VM %s stopped", vmID)
}

func queryVMHandler(w http.ResponseWriter, r *http.Request) {
	vmID := r.URL.Query().Get("vm_id")
	if vmID == "" {
		http.Error(w, "missing vm_id", 400)
		return
	}
	mu.RLock()
	inst, exists := instanceMap[vmID]
	mu.RUnlock()
	if !exists {
		http.Error(w, "vm not found", 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(inst)
}

func main() {
	if err := initDir(); err != nil {
		fmt.Println("init dir err:", err)
		return
	}
	http.HandleFunc("/vm/create", createVMHandler)
	http.HandleFunc("/vm/stop", stopVMHandler)
	http.HandleFunc("/vm/info", queryVMHandler)

	fmt.Println("Orchestrator running on", listenAddr)
	_ = http.ListenAndServe(listenAddr, nil)
}
