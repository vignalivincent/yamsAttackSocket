package api

import (
	"encoding/json"

	"github.com/vincentvignali/yamsAttackSocket/internal/game"
)


type GameHTTPHandler struct {
	gameManager *game.GameManager
}

type InitGameRequest struct {
	HostPlayerID string          `json:"hostPlayerId"`
	GameState    json.RawMessage `json:"gameState"`
}

type AppError struct {
	Code    int
	Message string
	Err     error
}