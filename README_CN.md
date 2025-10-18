# Mastersteam

[![GitHub release](https://img.shields.io/github/v/release/cyxc1124/Mastersteam)](https://github.com/cyxc1124/Mastersteam/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/cyxc1124/Mastersteam)](https://go.dev/)
[![License](https://img.shields.io/github/license/cyxc1124/Mastersteam)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/cyxc1124/mastersteam)](https://github.com/cyxc1124/Mastersteam/pkgs/container/mastersteam)

[English](README.md) | [中文](README_CN.md)

一个轻量级、高效的 Steam 游戏服务器查询 Web API。使用 Go 构建，提供 RESTful JSON 接口。

## ✨ 特性

- 🔥 **Steam Web API 集成** - 使用官方 Steam Web API，查询可靠稳定
- ⚡ **高性能** - 并发处理，可配置工作池
- 🎮 **通用支持** - 支持所有拥有专用服务器的 Steam 游戏
- 📊 **丰富的 JSON API** - RESTful 端点，提供完整的服务器信息
- 🐳 **Docker 就绪** - 预构建的多架构 Docker 镜像
- 🌍 **跨平台** - 提供 Windows、Linux 和 macOS 二进制文件
- 🔍 **灵活过滤** - 按 AppID、服务器名称或 IP 地址搜索
- 📈 **玩家信息** - 获取详细的玩家统计数据和在线玩家列表

## 📋 要求

- **Steam Web API 密钥**（必需）- 在 [steamcommunity.com/dev/apikey](https://steamcommunity.com/dev/apikey) 获取
- Go 1.24+ (如果从源码构建)
- Docker (可选，用于容器化部署)

> **注意：** 旧版 UDP Master Server (`hl2master.steampowered.com:27011`) 已弃用且不再可靠。此版本需要 Steam Web API 密钥。

## 🚀 快速开始

### 方式 1：Docker（推荐）

```bash
docker run -d \
  --name mastersteam \
  -p 8080:8080 \
  -e STEAM_API_KEY="你的API密钥" \
  ghcr.io/cyxc1124/mastersteam:latest
```

或使用 docker-compose：

```yaml
# docker-compose.yml
version: '3.8'
services:
  mastersteam:
    image: ghcr.io/cyxc1124/mastersteam:latest
    ports:
      - "8080:8080"
    environment:
      - STEAM_API_KEY=你的API密钥
    restart: unless-stopped
```

```bash
docker-compose up -d
```

### 方式 2：预编译二进制文件

从 [Releases](https://github.com/cyxc1124/Mastersteam/releases) 下载适合你平台的最新版本。

**Windows (PowerShell):**
```powershell
$env:STEAM_API_KEY="你的API密钥"
.\Mastersteam.exe
```

**Linux/macOS:**
```bash
export STEAM_API_KEY="你的API密钥"
./Mastersteam
```

### 方式 3：从源码构建

```bash
# 克隆仓库
git clone https://github.com/cyxc1124/Mastersteam.git
cd Mastersteam

# 构建
go build

# 运行
export STEAM_API_KEY="你的API密钥"
./Mastersteam
```

## 📖 API 文档

### 基础 URL
```
http://localhost:8080
```

### 端点

#### 1. 按 AppID 和名称搜索服务器

```http
GET /search/{APP_ID}/{NAME}
```

**参数：**
- `APP_ID` - Steam 应用程序 ID ([Steam App ID 列表](https://developer.valvesoftware.com/wiki/Steam_Application_IDs))
- `NAME` - 要搜索的服务器名称（使用 `*` 作为通配符）

**示例：**
```bash
# 搜索所有 CS:GO 服务器
curl "http://localhost:8080/search/730/*"

# 搜索名称中包含 "dust" 的 CS:GO 服务器
curl "http://localhost:8080/search/730/dust"

# 搜索 Arma 3 官方服务器
curl "http://localhost:8080/search/107410/*Bohemia%20Interactive*"
```

#### 2. 按 IP 地址搜索服务器

```http
GET /server/{IP}
```

**参数：**
- `IP` - 服务器 IP 地址（端口可选）

**示例：**
```bash
# 仅按 IP 搜索
curl "http://localhost:8080/server/192.168.1.1"

# 按 IP 和端口搜索
curl "http://localhost:8080/server/192.168.1.1:27015"
```

### 响应格式

```json
{
  "data": [{
    "192.168.1.1:27015": {
      "ip": "192.168.1.1:27015",
      "protocol": 17,
      "name": "我的超棒服务器",
      "map": "de_dust2",
      "folder": "csgo",
      "game": "Counter-Strike: Global Offensive",
      "players": 12,
      "max_players": 24,
      "bots": 0,
      "type": "dedicated",
      "os": "linux",
      "visibility": "public",
      "vac": true,
      "appid": 730,
      "game_version": "1.38.7.9",
      "port": 27015,
      "steamid": "90123456789012345",
      "game_mode": "casual",
      "gameid": "730",
      "players_online": [
        {
          "Name": "玩家1",
          "Score": 25,
          "Duration": 1234.56
        }
      ]
    }
  }],
  "total": 1
}
```

### 错误响应

```json
{
  "error": "Invalid Steam API Key",
  "details": "invalid API key (status 403): your Steam API key is invalid or expired",
  "status": 403
}
```

## 🛠️ 配置

### 环境变量

| 变量 | 必需 | 默认值 | 描述 |
|------|------|--------|------|
| `STEAM_API_KEY` | 是 | - | 你的 Steam Web API 密钥 |
| `PORT` | 否 | 8080 | HTTP 服务器端口 |

### 获取 Steam API 密钥

1. 访问 [steamcommunity.com/dev/apikey](https://steamcommunity.com/dev/apikey)
2. 使用你的 Steam 账号登录
3. 输入你的域名（本地开发使用 `localhost`）
4. 复制生成的 API 密钥

## 🏗️ 从源码构建

### 前置要求

- Go 1.24 或更高版本
- Git

### 构建命令

```bash
# 克隆仓库
git clone https://github.com/cyxc1124/Mastersteam.git
cd Mastersteam

# 安装依赖
go mod download

# 为当前平台构建
go build -o Mastersteam

# 为特定平台构建
GOOS=linux GOARCH=amd64 go build -o Mastersteam-linux-amd64
GOOS=windows GOARCH=amd64 go build -o Mastersteam-windows-amd64.exe
GOOS=darwin GOARCH=arm64 go build -o Mastersteam-darwin-arm64
```

## 🐳 Docker

### 从 GitHub Container Registry 拉取

```bash
docker pull ghcr.io/cyxc1124/mastersteam:latest
docker pull ghcr.io/cyxc1124/mastersteam:v1.0.0
```

### 本地构建 Docker 镜像

```bash
docker build -t mastersteam:local .
```

### 支持的平台

- `linux/amd64`
- `linux/arm64`

## 📚 常见 Steam App ID

| 游戏 | App ID |
|------|--------|
| 反恐精英：全球攻势 | 730 |
| 军团要塞 2 | 440 |
| 武装突袭 3 | 107410 |
| Rust | 252490 |
| Garry's Mod | 4000 |
| 求生之路 2 | 550 |
| 反恐精英：起源 | 240 |

完整列表请访问：[Steam Application IDs](https://developer.valvesoftware.com/wiki/Steam_Application_IDs)

## 🤝 贡献

欢迎贡献！请随时提交 Pull Request。

1. Fork 本仓库
2. 创建你的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交你的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开一个 Pull Request

## 📝 许可证

本项目采用 GNU General Public License v3.0 或更高版本授权 - 详见 [LICENSE](LICENSE) 文件。

## 🔗 资源

- [Steam Master Server Query Protocol](https://developer.valvesoftware.com/wiki/Master_Server_Query_Protocol)
- [Steam Server Queries](https://developer.valvesoftware.com/wiki/Server_queries)
- [Steam Web API 文档](https://partner.steamgames.com/doc/webapi)

## 🙏 致谢

本项目受以下项目启发：
- [alliedmodders/blaster](https://github.com/alliedmodders/blaster)
- [rumblefrog/go-a2s](https://github.com/rumblefrog/go-a2s)
- [TowelSoftware/Mastersteam](https://github.com/TowelSoftware/Mastersteam)

## 📧 支持

如果你有任何问题或疑问，请：
- 提交 [Issue](https://github.com/cyxc1124/Mastersteam/issues)
- 查看现有的 [讨论](https://github.com/cyxc1124/Mastersteam/discussions)

---

<sub>Made with ❤️ by <a href="https://github.com/cyxc1124">cyxc1124</a></sub>

