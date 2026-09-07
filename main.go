package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var (
	filePath     string
	uuidStr      string
	argoDomain   string
	argoAuth     string
	argoPort     string
	cfIP         string
	cfPort       string
	name         string
	subFilePath  string
	listFilePath string
	configPath   string
	bootLogPath  string
)

func loadDotEnv(envFile string) {
	file, err := os.Open(envFile)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func initEnvs() {
	loadDotEnv(".env")

	filePath = getEnv("FILE_PATH", ".cache")
	absPath, err := filepath.Abs(filePath)
	if err == nil {
		filePath = absPath
	}
	uuidStr = getEnv("UUID", "5520fab5-56d4-48cb-8156-e58b1cc18442")
	argoDomain = getEnv("ARGO_DOMAIN", "")
	argoAuth = getEnv("ARGO_AUTH", "")
	argoPort = getEnv("ARGO_PORT", "8001")
	cfIP = getEnv("CFIP", "saas.sin.fan")
	cfPort = getEnv("CFPORT", "443")
	name = getEnv("NAME", "")

	subFilePath = filepath.Join(filePath, "sub.txt")
	listFilePath = filepath.Join(filePath, "list.txt")
	configPath = filepath.Join(filePath, "config.json")
	bootLogPath = filepath.Join(filePath, "boot.log")
}

func createDirectory() {
	os.MkdirAll(filePath, 0755)
}

func cleanupOldFiles() {
	targets := []string{"web", "bot", "boot.log", "list.txt", "config.json", "tunnel.json", "tunnel.yml"}
	for _, target := range targets {
		p := filepath.Join(filePath, target)
		os.RemoveAll(p)
	}
}

func downloadFile(fileName, fileURL string) bool {
	dest := filepath.Join(filePath, fileName)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(fileURL)
	if err != nil {
		fmt.Printf("Download %s failed: %v\n", fileName, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Download %s failed with status: %s\n", fileName, resp.Status)
		return false
	}

	out, err := os.Create(dest)
	if err != nil {
		fmt.Printf("Create file %s failed: %v\n", fileName, err)
		return false
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(dest)
		fmt.Printf("Copy file %s failed: %v\n", fileName, err)
		return false
	}
	fmt.Printf("Downloaded %s successfully\n", fileName)
	return true
}

func authorizeFiles(files []string) {
	for _, f := range files {
		p := filepath.Join(filePath, f)
		os.Chmod(p, 0775)
	}
}

func setupArgo() {
	if argoAuth == "" || argoDomain == "" {
		return
	}
	if strings.Contains(argoAuth, "TunnelSecret") {
		os.WriteFile(filepath.Join(filePath, "tunnel.json"), []byte(argoAuth), 0644)
		parts := strings.Split(argoAuth, `"`)
		tunnelID := ""
		if len(parts) > 11 {
			tunnelID = parts[11]
		}
		yml := fmt.Sprintf(`tunnel: %s
credentials-file: %s
protocol: http2

ingress:
  - hostname: %s
    service: http://localhost:%s
    originRequest:
      noTLSVerify: true
  - service: http_status:404
`, tunnelID, filepath.Join(filePath, "tunnel.json"), argoDomain, argoPort)
		os.WriteFile(filepath.Join(filePath, "tunnel.yml"), []byte(yml), 0644)
	}
}

func generateSingboxConfig() {
	portNum := 8001
	fmt.Sscanf(argoPort, "%d", &portNum)

	config := map[string]interface{}{
		"log": map[string]interface{}{
			"disabled": true,
			"level":    "warn",
		},
		"inbounds": []interface{}{
			map[string]interface{}{
				"tag":         "vmess-ws-in",
				"type":        "vmess",
				"listen":      "::",
				"listen_port": portNum,
				"users": []interface{}{
					map[string]interface{}{
						"uuid": uuidStr,
					},
				},
				"transport": map[string]interface{}{
					"type":                   "ws",
					"path":                   "/vmess-argo",
					"early_data_header_name": "Sec-WebSocket-Protocol",
				},
			},
		},
		"outbounds": []interface{}{
			map[string]interface{}{
				"type": "direct",
				"tag":  "direct",
			},
		},
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(configPath, data, 0644)
}

func getArgoDomainFromLog() string {
	re := regexp.MustCompile(`https?://([^ ]*trycloudflare\.com)/?`)
	for i := 0; i < 15; i++ {
		time.Sleep(2 * time.Second)
		data, err := os.ReadFile(bootLogPath)
		if err == nil {
			matches := re.FindStringSubmatch(string(data))
			if len(matches) > 1 {
				return matches[1]
			}
		}
	}
	return ""
}

func generateNodes(domain string) string {
	nodeName := strings.TrimSpace(name)
	if nodeName == "" {
		nodeName = "Argo-Node"
	}

	cfPortNum := 443
	fmt.Sscanf(cfPort, "%d", &cfPortNum)

	vmessMap := map[string]interface{}{
		"v":    "2",
		"ps":   nodeName,
		"add":  cfIP,
		"port": cfPortNum,
		"id":   uuidStr,
		"aid":  "0",
		"scy":  "auto",
		"net":  "ws",
		"type": "none",
		"host": domain,
		"path": "/vmess-argo?ed=2560",
		"tls":  "tls",
		"sni":  domain,
		"alpn": "",
		"fp":   "firefox",
	}

	jsonBytes, _ := json.Marshal(vmessMap)
	vmessNode := "vmess://" + base64.StdEncoding.EncodeToString(jsonBytes)
	b64Sub := base64.StdEncoding.EncodeToString([]byte(vmessNode))

	os.WriteFile(subFilePath, []byte(b64Sub), 0644)
	os.WriteFile(listFilePath, []byte(vmessNode), 0644)

	fmt.Printf("\033[32m%s\033[0m\n", b64Sub)
	return b64Sub
}

func cleanFilesDelay() {
	time.AfterFunc(120*time.Second, func() {
		entries, err := os.ReadDir(filePath)
		if err != nil {
			return
		}
		for _, entry := range entries {
			if entry.Name() == "sub.txt" {
				continue
			}
			os.RemoveAll(filepath.Join(filePath, entry.Name()))
		}
		fmt.Printf("Cleanup done: preserved only %s\n", filepath.Join(filePath, "sub.txt"))
	})
}

func startServices() {
	fmt.Println("Starting ultra-minimal Argo tunnel service (Go native)...")
	cleanupOldFiles()
	createDirectory()

	arch := "amd"
	if strings.Contains(runtime.GOARCH, "arm") {
		arch = "arm"
	}
	baseURL := "https://amd64.ssss.nyc.mn"
	if arch == "arm" {
		baseURL = "https://arm64.ssss.nyc.mn"
	}

	webOK := downloadFile("web", baseURL+"/sb")
	botOK := downloadFile("bot", baseURL+"/2go")

	if !webOK || !botOK {
		fmt.Println("Failed to download required binaries (web/bot)")
		return
	}

	authorizeFiles([]string{"web", "bot"})
	setupArgo()
	generateSingboxConfig()

	// 1. Launch Sing-box (web)
	webBin := filepath.Join(filePath, "web")
	cmdWeb := exec.Command(webBin, "run", "-c", configPath)
	cmdWeb.Env = append(os.Environ(), "GOMEMLIMIT=12MiB", "GOGC=5")
	if err := cmdWeb.Start(); err != nil {
		fmt.Printf("Sing-box start failed: %v\n", err)
	} else {
		fmt.Println("Sing-box (web) started")
	}

	// 2. Launch Cloudflared (bot)
	targetDomain := argoDomain
	botBin := filepath.Join(filePath, "bot")
	var args []string
	if argoAuth != "" && argoDomain != "" {
		// 固定隧道模式
		fmt.Println("Mode: Fixed Tunnel")
		if strings.Contains(argoAuth, "TunnelSecret") {
			ymlPath := filepath.Join(filePath, "tunnel.yml")
			args = []string{"tunnel", "--edge-ip-version", "auto", "--config", ymlPath, "run"}
		} else {
			args = []string{"tunnel", "--edge-ip-version", "auto", "--no-autoupdate", "--protocol", "http2", "run", "--token", argoAuth}
		}
	} else {
		// 临时隧道模式
		fmt.Println("Mode: Temporary / Quick Tunnel")
		args = []string{"tunnel", "--edge-ip-version", "auto", "--no-autoupdate", "--protocol", "http2", "--logfile", bootLogPath, "--loglevel", "info", "--url", "http://localhost:" + argoPort}
	}

	cmdBot := exec.Command(botBin, args...)
	cmdBot.Env = append(os.Environ(), "GOMEMLIMIT=12MiB", "GOGC=5")
	if err := cmdBot.Start(); err != nil {
		fmt.Printf("Cloudflared start failed: %v\n", err)
	} else {
		fmt.Println("Cloudflared (bot) started")
	}

	if targetDomain == "" {
		targetDomain = getArgoDomainFromLog()
		fmt.Printf("Obtained temporary Argo domain: %s\n", targetDomain)
	}

	generateNodes(targetDomain)
	cleanFilesDelay()
}

func main() {
	initEnvs()
	startServices()
	// 维持 Go 守护进程持续运行
	for {
		time.Sleep(3600 * time.Second)
	}
}
