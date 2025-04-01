package game

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type HostConnectionState int

const (
	// HostNeverConnected : the host has never connected to the game yet
	HostNeverConnected HostConnectionState = iota
	// HostConnected : the host is currently connected
	HostConnected
	// HostDisconnected : the host was connected before but is currently disconnected
	HostDisconnected
)

type GameManager struct {
	games      map[string]*Game
	gamesMutex sync.Mutex
	Stats      *ServerStats
}

type Game struct {
	HostConnectionState HostConnectionState
	GameID              string          `json:"gameId"`
	HostPlayerID        string          `json:"hostPlayerId"`
	GameState           json.RawMessage `json:"gameState"`
	HostConn            *websocket.Conn
	Viewers             []*websocket.Conn
	Mutex               sync.Mutex
	CreatedAt           time.Time
	LastActivity        time.Time
}

type ServerStats struct {
	TotalGamesCreated    int
	ActiveGames          int
	TotalViewers         int
	TotalHostConnections int
	StartTime            time.Time
	Mutex                sync.RWMutex
}

type ServerStatsResponse struct {
	TotalGamesCreated    int    `json:"totalGamesCreated"`
	ActiveGames          int    `json:"activeGames"`
	TotalViewers         int    `json:"totalViewers"`
	TotalHostConnections int    `json:"totalHostConnections"`
	Uptime               string `json:"uptime"`
	StartTime            string `json:"startTime"`
}
