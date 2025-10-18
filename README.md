# Mastersteam

[![GitHub release](https://img.shields.io/github/v/release/cyxc1124/Mastersteam)](https://github.com/cyxc1124/Mastersteam/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/cyxc1124/Mastersteam)](https://go.dev/)
[![License](https://img.shields.io/github/license/cyxc1124/Mastersteam)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/cyxc1124/mastersteam)](https://github.com/cyxc1124/Mastersteam/pkgs/container/mastersteam)

[English](README.md) | [中文](README_CN.md)

A lightweight and efficient web API for querying game servers from the Steam server list. Built with Go, providing RESTful JSON endpoints.

## ✨ Features

- 🔥 **Steam Web API Integration** - Uses official Steam Web API for reliable server queries
- ⚡ **High Performance** - Concurrent processing with configurable worker pools
- 🎮 **Universal Support** - Works with all Steam games that have dedicated servers
- 📊 **Rich JSON API** - RESTful endpoints with comprehensive server information
- 🐳 **Docker Ready** - Pre-built multi-architecture Docker images
- 🌍 **Cross-Platform** - Binaries available for Windows, Linux, and macOS
- 🔍 **Flexible Filtering** - Search by AppID, server name, or IP address
- 📈 **Player Information** - Get detailed player stats and online player lists

## 📋 Requirements

- **Steam Web API Key** (Required) - Get yours at [steamcommunity.com/dev/apikey](https://steamcommunity.com/dev/apikey)
- Go 1.24+ (for building from source)
- Docker (optional, for containerized deployment)

> **Note:** The legacy UDP Master Server (`hl2master.steampowered.com:27011`) is deprecated and no longer reliable. This version requires a Steam Web API key.

## 🚀 Quick Start

### Option 1: Docker (Recommended)

```bash
docker run -d \
  --name mastersteam \
  -p 8080:8080 \
  -e STEAM_API_KEY="YOUR_API_KEY_HERE" \
  ghcr.io/cyxc1124/mastersteam:latest
```

Or using docker-compose:

```yaml
# docker-compose.yml
version: '3.8'
services:
  mastersteam:
    image: ghcr.io/cyxc1124/mastersteam:latest
    ports:
      - "8080:8080"
    environment:
      - STEAM_API_KEY=YOUR_API_KEY_HERE
    restart: unless-stopped
```

```bash
docker-compose up -d
```

### Option 2: Pre-built Binaries

Download the latest release for your platform from [Releases](https://github.com/cyxc1124/Mastersteam/releases).

**Windows (PowerShell):**
```powershell
$env:STEAM_API_KEY="YOUR_API_KEY_HERE"
.\Mastersteam.exe
```

**Linux/macOS:**
```bash
export STEAM_API_KEY="YOUR_API_KEY_HERE"
./Mastersteam
```

### Option 3: Build from Source

```bash
# Clone the repository
git clone https://github.com/cyxc1124/Mastersteam.git
cd Mastersteam

# Build
go build

# Run
export STEAM_API_KEY="YOUR_API_KEY_HERE"
./Mastersteam
```

## 📖 API Documentation

### Base URL
```
http://localhost:8080
```

### Endpoints

#### 1. Search Servers by AppID and Name

```http
GET /search/{APP_ID}/{NAME}
```

**Parameters:**
- `APP_ID` - Steam Application ID ([List of Steam App IDs](https://developer.valvesoftware.com/wiki/Steam_Application_IDs))
- `NAME` - Server name to search (use `*` for wildcard)

**Example:**
```bash
# Search all CS:GO servers
curl "http://localhost:8080/search/730/*"

# Search CS:GO servers with "dust" in name
curl "http://localhost:8080/search/730/dust"

# Search Arma 3 official servers
curl "http://localhost:8080/search/107410/*Bohemia%20Interactive*"
```

#### 2. Search Servers by IP Address

```http
GET /server/{IP}
```

**Parameters:**
- `IP` - Server IP address (port optional)

**Example:**
```bash
# Search by IP only
curl "http://localhost:8080/server/192.168.1.1"

# Search by IP and port
curl "http://localhost:8080/server/192.168.1.1:27015"
```

### Response Format

```json
{
  "data": [{
    "192.168.1.1:27015": {
      "ip": "192.168.1.1:27015",
      "protocol": 17,
      "name": "My Awesome Server",
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
          "Name": "Player1",
          "Score": 25,
          "Duration": 1234.56
        }
      ]
    }
  }],
  "total": 1
}
```

### Error Response

```json
{
  "error": "Invalid Steam API Key",
  "details": "invalid API key (status 403): your Steam API key is invalid or expired",
  "status": 403
}
```

## 🛠️ Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `STEAM_API_KEY` | Yes | - | Your Steam Web API key |
| `PORT` | No | 8080 | HTTP server port |

### Getting a Steam API Key

1. Visit [steamcommunity.com/dev/apikey](https://steamcommunity.com/dev/apikey)
2. Log in with your Steam account
3. Enter your domain name (use `localhost` for local development)
4. Copy the generated API key

## 🏗️ Building from Source

### Prerequisites

- Go 1.24 or higher
- Git

### Build Commands

```bash
# Clone repository
git clone https://github.com/cyxc1124/Mastersteam.git
cd Mastersteam

# Install dependencies
go mod download

# Build for current platform
go build -o Mastersteam

# Build for specific platforms
GOOS=linux GOARCH=amd64 go build -o Mastersteam-linux-amd64
GOOS=windows GOARCH=amd64 go build -o Mastersteam-windows-amd64.exe
GOOS=darwin GOARCH=arm64 go build -o Mastersteam-darwin-arm64
```

## 🐳 Docker

### Pull from GitHub Container Registry

```bash
docker pull ghcr.io/cyxc1124/mastersteam:latest
docker pull ghcr.io/cyxc1124/mastersteam:v1.0.0
```

### Build Docker Image Locally

```bash
docker build -t mastersteam:local .
```

### Supported Platforms

- `linux/amd64`
- `linux/arm64`

## 📚 Common Steam App IDs

| Game | App ID |
|------|--------|
| Counter-Strike: Global Offensive | 730 |
| Team Fortress 2 | 440 |
| Arma 3 | 107410 |
| Rust | 252490 |
| Garry's Mod | 4000 |
| Left 4 Dead 2 | 550 |
| Counter-Strike: Source | 240 |

For a complete list, visit: [Steam Application IDs](https://developer.valvesoftware.com/wiki/Steam_Application_IDs)

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the GNU General Public License v3.0 or higher - see the [LICENSE](LICENSE) file for details.

## 🔗 Resources

- [Steam Master Server Query Protocol](https://developer.valvesoftware.com/wiki/Master_Server_Query_Protocol)
- [Steam Server Queries](https://developer.valvesoftware.com/wiki/Server_queries)
- [Steam Web API Documentation](https://partner.steamgames.com/doc/webapi)

## 🙏 Credits

This project was inspired by:
- [alliedmodders/blaster](https://github.com/alliedmodders/blaster)
- [rumblefrog/go-a2s](https://github.com/rumblefrog/go-a2s)
- [TowelSoftware/Mastersteam](https://github.com/TowelSoftware/Mastersteam)

## 📧 Support

If you have any questions or issues, please:
- Open an [Issue](https://github.com/cyxc1124/Mastersteam/issues)
- Check existing [Discussions](https://github.com/cyxc1124/Mastersteam/discussions)

---

<sub>Made with ❤️ by <a href="https://github.com/cyxc1124">cyxc1124</a></sub>
