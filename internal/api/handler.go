/*
HTTP API Layer

This file implements the GameHTTPHandler component, which manages the HTTP API endpoints
for the game sharing service. It serves as the primary interface for clients to create
shared game sessions and retrieve server statistics.

Key responsibilities:
- Creating new shared game sessions with initial game state
- Providing server metrics and statistics
- Validating incoming requests and parameters
- Generating proper HTTP responses with appropriate status codes
- Handling errors in a consistent way across the API

The GameHTTPHandler interfaces with the GameManager to create games and retrieve metrics,
providing a RESTful API layer that complements the WebSocket-based real-time communication.
It implements a clean separation between the HTTP concerns and the underlying game management
logic, making the codebase more maintainable and testable.

This component is essential for establishing game sessions that can then be accessed
through the WebSocket connections for real-time updates.
*/

package api

import (
	"encoding/json"
	"net/http"

	"github.com/vincentvignali/yamsAttackSocket/internal/game"
)


func NewGameHTTPHandler(gameManager *game.GameManager) *GameHTTPHandler {
	return &GameHTTPHandler{
		gameManager: gameManager,
	}
}

func (h *GameHTTPHandler) InitSharedGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HandleError(w, &AppError{
			Code:    http.StatusMethodNotAllowed,
			Message: ErrMethodNotAllowed,
		})
		return
	}

	if r.Body == nil || r.ContentLength == 0 {
		HandleError(w, &AppError{
			Code:    http.StatusBadRequest,
			Message: ErrNoBody,
		})
		return
	}

	var req InitGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		HandleError(w, &AppError{
			Code:    http.StatusBadRequest,
			Message: ErrJSONParsing,
			Err:     err,
		})
		return
	}
	
	if req.HostPlayerID == "" {
		HandleError(w, &AppError{
			Code:    http.StatusBadRequest,
			Message: ErrMissingParam + ": hostPlayerId",
		})
		return
	}
	if len(req.GameState) == 0 {
		HandleError(w, &AppError{
			Code:    http.StatusBadRequest,
			Message: ErrMissingParam + ": gameState",
		})
		return
	}

	gameID, err := h.gameManager.CreateGame(req.HostPlayerID, req.GameState)
	if err != nil {
		HandleError(w, &AppError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create game",
			Err:     err,
		})
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	baseURL := scheme + "://" + r.Host
	shareURL := baseURL + "?viewer=true&gameId=" + gameID

	response := InitGameResponse{
		GameID:   gameID,
		ShareURL: shareURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *GameHTTPHandler) ServerStats(w http.ResponseWriter, r *http.Request) {
	metrics := h.gameManager.GetMetrics()
	

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics.FormatResponse())
}