# 🚀 Sing-box Argo Python 极简低内存节点服务 (100M 容器开源专版)

## ⭐ Star 一下支持项目 ⭐

本项目是一个专为 **100M 限制低内存容器**（如 Pella 等）打造的轻量级 VMess-WS + Argo 隧道代理节点托管服务。

采用 **“Go 语言原生机器码 + 超轻量 Python Gunicorn 引导”** 架构，将容器运行总内存降至 **25MB ~ 40MB**，即使在节点大流量并发传输时也绝对不会产生 OOM（内存爆满崩溃重启）。

---

## ✨ 项目核心特点

1. **极致内存优化 (25MB ~ 40MB)**
   - 核心代理与进程控制使用 Go 语言原生实现，去除了 Python 虚拟机与庞大框架的 ~25MB 硬性基准内存开销。
   - 彻底移除了 WARP (WireGuard) 路由和远程 Geosite 规则下载，仅保留极简 `vmess-ws-in` 与 `direct` 出站。

2. **双模式隧道自动切换 (固定隧道 vs 临时隧道)**
   - **固定隧道**：填写 `ARGO_DOMAIN` 与 `ARGO_AUTH` 环境变量，自动建立 Cloudflare 自定义固定隧道。
   - **临时隧道**：将 `ARGO_DOMAIN` 和 `ARGO_AUTH` **留空**，程序自动发起 Cloudflare Quick Tunnel 并在日志中捕获生成免费的 `trycloudflare.com` 临时节点！
   - **提示：建议使用固定隧道，临时隧道在容器重启后需重新申请隧道，隧道域名会有变化，需重新导入节点**

3. **100% 兼容 Gunicorn / WSGI 托管平台**
   - 附带 30 行超轻量 `main.py` 入口，完美通过 Gunicorn/Pella 平台的 `importlib` 健康检查，并在后台自动静默拉起 Go 二进制 `./main`。

4. **日志隐私与自动清理**
   - 节点输出为加密 Base64 订阅密文（以 `dm1lc3M6...` 开头），不暴露出明文链接。
   - 启动 2 分钟后自动清理运行目录中的二进制文件与配置文件，仅保留 `sub.txt` 供订阅路由读取，极大释放磁盘与内存。

---

## ✨ 账号注册注意事项

注册网址：https://www.pella.app/

点击 SIGN UP 注册

**注册账号注意：谷歌Gmail邮箱使用邮箱+密码注册，续期脚本要用密码，不要使用关联账号注册，一旦注册，不可更改，不可删除**

登录页面后选择：Web App 分类，，接着选择 Flask 类型，这个就是python架构，，接着源代码上传可以暂时留空，继续Continue，，最后选free计划，开启容器，上传本项目文件，main二进制文件较大，请耐心等待，运行容器

续期方法前往该项目：https://github.com/liuhuig/PellaFree-pro

---

## 🛠️ 环境变量配置说明

所有环境变量在 [main.py](main.py) 顶部的 `os.environ.setdefault(...)` 中直接编辑默认值。

| 环境变量 | 默认值 | 说明 |
| :--- | :--- | :--- |
| **`FILE_PATH`** | `.cache` | 运行路径，sub.txt 保存目录 |
| **`UUID`** | `5520fab5-56d4-48cb-8156-e58b1cc18442` | VMess 用户 UUID |
| **`ARGO_DOMAIN`** | `""` | 固定隧道域名 |
| **`ARGO_AUTH`** | `""` | 固定隧道 Token |
| **`ARGO_PORT`** | `8001` | Argo 隧道本地监听端口 |
| **`CFIP`** | `saas.sin.fan` | 优选 IP 或优选域名 |
| **`CFPORT`** | `443` | 优选端口 |
| **`NAME`** | `""` | 节点显示名称 |


---

## 🚀 快速部署步骤

### 方案 A：上传项目文件部署（推荐）
1. 将打包好的 **`main.py`**、**`main`**（可执行二进制）和 **`requirements.txt`** 上传至容器，上传时间较长，请耐心等待
2. 保持平台默认的启动命令不变（`main.py`）。
3. 容器启动后，日志输出节点信息，同时在文件夹目录.cache下保存sub.txt

---

## 🔧 自行源码编译指南 (开发者)

如果您修改了 `main.go` 源码，需要在本地跨平台编译为 Linux 64 位 ELF 二进制文件：

### 在 Windows (PowerShell) 下编译：
```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -ldflags="-s -w" -o main main.go
```

### 在 Linux / macOS 下编译：
```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o main main.go
```

---

## 📄 开源协议
本项目采用 [MIT License](LICENSE) 协议开源。

