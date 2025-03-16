package websocket

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/vincentvignali/yamsAttackSocket/internal/game"
)

type GameWSHandler struct {
	gameManager *game.GameManager
	upgrader    websocket.Upgrader
}

type HostMessage struct {
	GameState json.RawMessage `json:"gameState"`
}