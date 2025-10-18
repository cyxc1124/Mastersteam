// Licensed under the GNU General Public License, version 3 or higher.
package valve

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SteamWebAPIQuerier queries server lists using Steam Web API
type SteamWebAPIQuerier struct {
	apiKey  string
	client  *http.Client
	filters []string
}

// steamWebAPIResponse is the response structure from Steam Web API
type steamWebAPIResponse struct {
	Response struct {
		Servers []struct {
			Addr       string `json:"addr"`
			Gameport   int    `json:"gameport"`
			Steamid    string `json:"steamid"`
			Name       string `json:"name"`
			Appid      int    `json:"appid"`
			Gamedir    string `json:"gamedir"`
			Version    string `json:"version"`
			Product    string `json:"product"`
			Region     int    `json:"region"`
			Players    int    `json:"players"`
			MaxPlayers int    `json:"max_players"`
			Bots       int    `json:"bots"`
			Map        string `json:"map"`
			Secure     bool   `json:"secure"`
			Dedicated  bool   `json:"dedicated"`
			Os         string `json:"os"`
			GameType   string `json:"gametype"`
		} `json:"servers"`
	} `json:"response"`
}

// NewSteamWebAPIQuerier creates a new Steam Web API querier
func NewSteamWebAPIQuerier(apiKey string) (*SteamWebAPIQuerier, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("steam API key is required")
	}

	return &SteamWebAPIQuerier{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: time.Minute * 2,
		},
	}, nil
}

// FilterAppId adds an AppID filter
func (q *SteamWebAPIQuerier) FilterAppId(appId AppId) {
	q.filters = append(q.filters, fmt.Sprintf("appid\\%d", appId))
}

// FilterAppIds adds multiple AppID filters
func (q *SteamWebAPIQuerier) FilterAppIds(appIds []AppId) {
	for _, appId := range appIds {
		q.FilterAppId(appId)
	}
}

// FilterName adds a server name filter
func (q *SteamWebAPIQuerier) FilterName(serverName string) {
	if serverName != "" && serverName != "*" {
		q.filters = append(q.filters, fmt.Sprintf("name_match\\%s", serverName))
	}
}

// FilterGameaddr adds an IP address filter
func (q *SteamWebAPIQuerier) FilterGameaddr(serverIP string) {
	if serverIP != "" {
		q.filters = append(q.filters, fmt.Sprintf("gameaddr\\%s", serverIP))
	}
}

// Query queries the server list
func (q *SteamWebAPIQuerier) Query(callback MasterQueryCallback) error {
	// Build filter string
	filterStr := q.buildFilterString()

	// Build API URL
	apiURL := fmt.Sprintf(
		"%s?key=%s&filter=%s&limit=10000",
		SteamWebAPIURL,
		q.apiKey,
		url.QueryEscape(filterStr),
	)

	// Send HTTP request
	resp, err := q.client.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to query Steam Web API: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		// Provide specific error messages based on status code
		switch resp.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			return fmt.Errorf("invalid API key (status %d): your Steam API key is invalid or expired, get a new key from https://steamcommunity.com/dev/apikey", resp.StatusCode)
		case http.StatusTooManyRequests:
			return fmt.Errorf("rate limit exceeded (status 429): too many requests to Steam API, please wait and try again")
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			return fmt.Errorf("Steam API service error (status %d): Steam servers may be down or experiencing issues", resp.StatusCode)
		default:
			return fmt.Errorf("Steam Web API error (status %d): %s", resp.StatusCode, string(body))
		}
	}

	// Parse JSON response
	var result steamWebAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode Steam Web API response: %w", err)
	}

	// Convert to ServerList format
	servers := make(ServerList, 0, len(result.Response.Servers))
	for _, srv := range result.Response.Servers {
		// Parse server address
		addr, err := net.ResolveTCPAddr("tcp", srv.Addr)
		if err != nil {
			// If address parsing fails, try to construct manually
			parts := strings.Split(srv.Addr, ":")
			if len(parts) == 2 {
				addr = &net.TCPAddr{
					IP:   net.ParseIP(parts[0]),
					Port: srv.Gameport,
				}
			}
			if addr == nil || addr.IP == nil {
				continue // Skip invalid addresses
			}
		}
		servers = append(servers, addr)
	}

	// Call callback function
	if len(servers) > 0 {
		return callback(servers)
	}

	return nil
}

// buildFilterString builds the filter string
func (q *SteamWebAPIQuerier) buildFilterString() string {
	if len(q.filters) == 0 {
		return ""
	}

	// Combine all filters into a single string
	// Format: \appid\730\name_match\test
	var builder strings.Builder
	for _, filter := range q.filters {
		builder.WriteString("\\")
		builder.WriteString(filter)
	}

	return builder.String()
}

// Close closes the connection (Web API doesn't need to close, kept for interface compatibility)
func (q *SteamWebAPIQuerier) Close() {
	// HTTP client manages connection pool automatically
}
