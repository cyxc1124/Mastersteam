// Licensed under the GNU General Public License, version 3 or higher.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	batch "github.com/cyxc1124/Mastersteam/batch"
	valve "github.com/cyxc1124/Mastersteam/valve"
)

// Version information (set by build flags)
var (
	GitTag      = "dev"
	GitCommit   = "unknown"
	GitBranch   = "unknown"
	BuildTime   = "unknown"
	BuildNumber = "0"
)

var (
	sOutputBuffer bytes.Buffer
	sNumServers   int64
	master        valve.MasterQuerier
)

/*
ErrorObject ...
*/
type ErrorObject struct {
	IP    string `json:"ip"`
	Error string `json:"error"`
}

/*
ServerObject ...
*/
type ServerObject struct {
	Address     string      `json:"ip"`
	Protocol    uint8       `json:"protocol"`
	Name        string      `json:"name"`
	MapName     string      `json:"map"`
	Folder      string      `json:"folder"`
	Game        string      `json:"game"`
	Players     uint8       `json:"players"`
	MaxPlayers  uint8       `json:"max_players"`
	Bots        uint8       `json:"bots"`
	Type        string      `json:"type"`
	Os          string      `json:"os"`
	Visibility  string      `json:"visibility"`
	Vac         bool        `json:"vac"`
	AppID       valve.AppId `json:"appid,omitempty"`
	GameVersion string      `json:"game_version,omitempty"`
	Port        uint16      `json:"port,omitempty"`
	SteamID     string      `json:"steamid,omitempty"`
	GameMode    string      `json:"game_mode,omitempty"`
	GameID      string      `json:"gameid,omitempty"`

	PlayersOnline []*valve.Player `json:"players_online,omitempty"`
}

func addJSON(hostAndPort string, obj interface{}) {
	buf, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	var indented bytes.Buffer
	json.Indent(&indented, buf, "\t", "\t")

	if sNumServers != 0 {
		sOutputBuffer.WriteString(",")
	}

	sOutputBuffer.WriteString(fmt.Sprintf("\n\t\"%s\": ", hostAndPort))
	sOutputBuffer.WriteString(indented.String())

	sNumServers++
}

func addError(hostAndPort string, err error) {
	// 记录详细错误到日志
	log.Printf("⚠️  Server query error [%s]: %s", hostAndPort, err.Error())

	// 只返回通用错误消息给用户，不暴露敏感信息
	userFriendlyError := "Query failed"
	if strings.Contains(err.Error(), "timeout") {
		userFriendlyError = "Connection timeout"
	} else if strings.Contains(err.Error(), "connection refused") {
		userFriendlyError = "Connection refused"
	} else if strings.Contains(err.Error(), "no route to host") {
		userFriendlyError = "Host unreachable"
	}

	addJSON(hostAndPort, &ErrorObject{
		IP:    hostAndPort,
		Error: userFriendlyError,
	})
}

/*
Log ...
*/
func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("access: %s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

// handleQueryError handles query errors and returns appropriate JSON response
func handleQueryError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	errMsg := err.Error()
	statusCode := http.StatusInternalServerError
	userMessage := "Failed to query server list"

	// Check for specific error types
	if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "403") || strings.Contains(errMsg, "Unauthorized") {
		statusCode = http.StatusUnauthorized
		userMessage = "Invalid Steam API Key"
		log.Printf("⚠️  ERROR: Invalid API Key - Please check your STEAM_API_KEY")
		log.Printf("   Get a valid key from: https://steamcommunity.com/dev/apikey")
	} else if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "EOF") {
		statusCode = http.StatusGatewayTimeout
		userMessage = "Steam API request timeout or connection error"
		log.Printf("⚠️  ERROR: Network error - %s", errMsg)
	} else if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "connection refused") {
		statusCode = http.StatusServiceUnavailable
		userMessage = "Cannot connect to Steam API"
		log.Printf("⚠️  ERROR: Connection error - %s", errMsg)
	} else {
		log.Printf("⚠️  ERROR: Query failed - %s", errMsg)
	}

	w.WriteHeader(statusCode)

	// 只返回友好的错误消息给用户，不包含敏感的技术细节
	errorResponse := map[string]interface{}{
		"error":  userMessage,
		"status": statusCode,
	}

	json.NewEncoder(w).Encode(errorResponse)
}

func httpMasterSearch(w http.ResponseWriter, r *http.Request) {
	uriSegments := strings.Split(r.URL.String(), "/")
	appID, _ := strconv.Atoi(uriSegments[2])
	hostname, _ := url.QueryUnescape(uriSegments[3])

	newWebAPIQuerier()

	// Set up the filter list.
	master.FilterAppId(valve.AppId(appID))
	master.FilterName(hostname)

	err := newServerQuerier()
	if err != nil {
		handleQueryError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	fmt.Fprintf(w, "%s", sOutputBuffer.String())
}

func httpServer(w http.ResponseWriter, r *http.Request) {
	uriSegments := strings.Split(r.URL.String(), "/")
	host, _ := url.QueryUnescape(uriSegments[2])

	newWebAPIQuerier()

	master.FilterGameaddr(host)

	err := newServerQuerier()
	if err != nil {
		handleQueryError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	fmt.Fprintf(w, "%s", sOutputBuffer.String())
}

func newWebAPIQuerier() {
	// Create Steam Web API querier
	m, err := valve.NewSteamWebAPIQuerier(valve.SteamAPIKey)
	if err != nil {
		log.Printf("ERROR: Failed to create Steam Web API querier: %s", err.Error())
		return
	}
	master = m
}

func newServerQuerier() error {
	flagTimeout := time.Second * 3
	flagJ := 20
	sNumServers = 0

	sOutputBuffer.Reset()

	bp := batch.NewBatchProcessor(func(item interface{}) {
		addr := item.(*net.TCPAddr)
		query, err := valve.NewServerQuerier(addr.String(), flagTimeout)
		if err != nil {
			addError(addr.String(), err)
			return
		}
		defer query.Close()

		info, err := query.QueryInfo()
		if err != nil {
			addError(addr.String(), err)
			return
		}

		log.Printf("%s - %s\n", addr.String(), info.Name)

		out := &ServerObject{
			Address:    addr.String(),
			Protocol:   info.Protocol,
			Name:       info.Name,
			MapName:    info.MapName,
			Folder:     info.Folder,
			Game:       info.Game,
			Players:    info.Players,
			MaxPlayers: info.MaxPlayers,
			Bots:       info.Bots,
			Type:       info.Type.String(),
			Os:         info.OS.String(),
		}
		if info.Vac == 1 {
			out.Vac = true
		}
		if info.Visibility == 0 {
			out.Visibility = "public"
		} else {
			out.Visibility = "private"
		}
		if info.Ext != nil {
			out.AppID = info.Ext.AppId
			out.GameVersion = info.Ext.GameVersion
			out.Port = info.Ext.Port
			out.SteamID = fmt.Sprintf("%d", info.Ext.SteamId)
			out.GameMode = info.Ext.GameModeDescription
			out.GameID = fmt.Sprintf("%d", info.Ext.GameId)
		}

		if info.Players > 0 {
			players, err := query.QueryPlayers()
			if err != nil {
				out.PlayersOnline = nil
			} else {
				out.PlayersOnline = players
			}
		}

		addJSON(addr.String(), out)
	}, flagJ)

	defer bp.Terminate()

	// TOP OF JSON FILE
	sOutputBuffer.WriteString("{\n")
	sOutputBuffer.WriteString("\t\"data\" : [{")

	// Query the master.
	err := master.Query(func(servers valve.ServerList) error {
		bp.AddBatch(servers)
		return nil
	})

	if err != nil {
		log.Printf("Failed to query server list: %s\n", err.Error())
		return err
	}

	// Wait for batch processing to complete.
	bp.Finish()

	sOutputBuffer.WriteString("}],\n")
	sOutputBuffer.WriteString(fmt.Sprintf("\t\"total\":%d\n", sNumServers))
	sOutputBuffer.WriteString("}\n")
	//BOTTOM OF JSON FILE

	return nil
}

func init() {
	// Read Steam API Key from environment variable
	valve.SteamAPIKey = os.Getenv("STEAM_API_KEY")

	if valve.SteamAPIKey == "" {
		log.Printf("⚠️  ERROR: STEAM_API_KEY environment variable not set")
		log.Printf("")
		log.Printf("This service requires a Steam Web API Key to function")
		log.Printf("")
		log.Printf("Get your API Key:")
		log.Printf("  1. Visit: https://steamcommunity.com/dev/apikey")
		log.Printf("  2. Log in with your Steam account")
		log.Printf("  3. Domain: localhost")
		log.Printf("  4. Copy the generated API Key")
		log.Printf("")
		log.Printf("Set the key:")
		log.Printf("  PowerShell:  $env:STEAM_API_KEY=\"YOUR_KEY\"")
		log.Printf("  Linux/macOS: export STEAM_API_KEY=\"YOUR_KEY\"")
		log.Printf("")
		log.Fatal("Cannot start service: Missing STEAM_API_KEY")
	}

	log.Printf("✓ Steam API Key configured")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	log.Printf("🚀 Mastersteam service starting")
	log.Printf("   Version: %s", GitTag)
	log.Printf("   Commit: %s", GitCommit)
	log.Printf("   Build Time: %s", BuildTime)
	log.Printf("   Query mode: Steam Web API")
	log.Printf("   Listening on port: 8080")
	log.Printf("")
	log.Printf("API Endpoints:")
	log.Printf("   GET /search/[APP_ID]/[NAME]")
	log.Printf("   GET /server/[IP]")
	log.Printf("")

	http.HandleFunc("/search/", httpMasterSearch)
	http.HandleFunc("/server/", httpServer)
	log.Fatal(http.ListenAndServe(":8080", Log(http.DefaultServeMux)))
}
