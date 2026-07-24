package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func GetSystemInfo() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	username := getCurrentUsername()
	ramBytes := getTotalRAMBytes()
	info := map[string]any{
		"hostname":  hostname,
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
		"user":      username,
		"ip":        getLocalIP(),
		"cpu":       getCPUModel(),
		"gpu":       getGPUModel(),
		"ram_bytes": ramBytes,
		"ram_human": formatBytes(ramBytes),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	fileName := fmt.Sprintf("%s_systemlog.json", getLocalIP())
	jsonPayload, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Printf("Failed to encode system info as JSON: %v", err)
		return
	}

	if err := os.WriteFile(fileName, jsonPayload, 0o644); err != nil {
		log.Printf("Failed to write %s: %v", fileName, err)
		return
	}


	uploadURL := fmt.Sprintf("http://%s:%s/api/upload", serverAddress, apiPort)
	if err := uploadFile(uploadURL, fileName); err != nil {
		log.Printf("Failed to upload %s to %s: %v", fileName, uploadURL, err)
		return
	}

	if err := os.Remove(fileName); err != nil {
		log.Printf("Uploaded %s but failed to delete local copy: %v", fileName, err)
		return
	}

	log.Printf("System information written to %s, uploaded to %s, and deleted locally", fileName, uploadURL)
}

func getLocalIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch value := addr.(type) {
			case *net.IPNet:
				ip = value.IP
			case *net.IPAddr:
				ip = value.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue
			}

			return ip.String()
		}
	}

	return "unknown"
}

func getCurrentUsername() string {
	currentUser, err := user.Current()
	if err == nil && currentUser.Username != "" {
		return currentUser.Username
	}

	if name := strings.TrimSpace(os.Getenv("USER")); name != "" {
		return name
	}
	if name := strings.TrimSpace(os.Getenv("USERNAME")); name != "" {
		return name
	}

	return "unknown"
}

func getCPUModel() string {
	switch runtime.GOOS {
	case "linux":
		if out := runCommand("sh", "-c", "lscpu | awk -F: '/Model name/ {gsub(/^ +/, \"\", $2); print $2; exit}'"); out != "" {
			return out
		}
		if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "model name") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						return strings.TrimSpace(parts[1])
					}
				}
			}
		}
	case "darwin":
		if out := runCommand("sysctl", "-n", "machdep.cpu.brand_string"); out != "" {
			return out
		}
	case "windows":
		if out := runCommand("wmic", "cpu", "get", "name", "/value"); out != "" {
			for _, line := range strings.Split(out, "\n") {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(strings.ToLower(trimmed), "name=") {
					return strings.TrimSpace(strings.TrimPrefix(trimmed, "Name="))
				}
			}
		}
	}

	return "unknown"
}

func getGPUModel() string {
	switch runtime.GOOS {
	case "linux":
		if out := runCommand("sh", "-c", "nvidia-smi --query-gpu=name --format=csv,noheader | head -n 1"); out != "" {
			return out
		}
		if out := runCommand("sh", "-c", "lspci | grep -Ei 'vga|3d|2d' | head -n 1"); out != "" {
			return out
		}
	case "darwin":
		if out := runCommand("sh", "-c", "system_profiler SPDisplaysDataType | awk -F: '/Chipset Model/ {gsub(/^ +/, \"\", $2); print $2; exit}'"); out != "" {
			return out
		}
	case "windows":
		if out := runCommand("wmic", "path", "win32_VideoController", "get", "name"); out != "" {
			for _, line := range strings.Split(out, "\n") {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" && !strings.EqualFold(trimmed, "name") {
					return trimmed
				}
			}
		}
	}

	return "unknown"
}

func getTotalRAMBytes() uint64 {
	switch runtime.GOOS {
	case "linux":
		if data, err := os.ReadFile("/proc/meminfo"); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "MemTotal:") {
					fields := strings.Fields(line)
					if len(fields) >= 2 {
						kb, parseErr := strconv.ParseUint(fields[1], 10, 64)
						if parseErr == nil {
							return kb * 1024
						}
					}
				}
			}
		}
	case "darwin":
		if out := runCommand("sysctl", "-n", "hw.memsize"); out != "" {
			value, err := strconv.ParseUint(strings.TrimSpace(out), 10, 64)
			if err == nil {
				return value
			}
		}
	case "windows":
		if out := runCommand("wmic", "computersystem", "get", "TotalPhysicalMemory", "/value"); out != "" {
			for _, line := range strings.Split(out, "\n") {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(strings.ToLower(trimmed), "totalphysicalmemory=") {
					value := strings.TrimSpace(strings.TrimPrefix(trimmed, "TotalPhysicalMemory="))
					parsed, err := strconv.ParseUint(value, 10, 64)
					if err == nil {
						return parsed
					}
				}
			}
		}
	}

	return 0
}

func formatBytes(b uint64) string {
	if b == 0 {
		return "unknown"
	}

	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func uploadFile(uploadURL string, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}

	if _, err := io.Copy(part, file); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, uploadURL, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	httpClient := &http.Client{Timeout: 15 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	return nil
}

func runCommand(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}
