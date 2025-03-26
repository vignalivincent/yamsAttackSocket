package websocket

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vincentvignali/yamsAttackSocket/internal/api"
	"github.com/vincentvignali/yamsAttackSocket/internal/game"
	"github.com/vincentvignali/yamsAttackSocket/internal/logger"
)

/*
WebSocket Communication Layer

This file implements the GameWSHandler component, which manages real-time bidirectional
communication between clients and the game server using WebSockets. It serves as the
communication layer between the frontend clients and the game state management system.

Key responsibilities:
- Establishing and maintaining WebSocket connections
- Authenticating host and viewer connections
- Managing real-time game state broadcasts
- Handling connection lifecycle (connect, disconnect, cleanup)
- Broadcasting game state updates from hosts to viewers
- Processing connection events and maintaining connection state

The GameWSHandler interfaces with the GameManager to access and modify game instances,
while providing real-time communication capabilities that complement the HTTP-based API.
It maintains separate handling logic for game hosts (who can update game state) and
viewers (who only receive updates).

This component is critical for enabling the live-sharing functionality that allows
players to share their game progress in real time with spectators.
*/



func NewGameWSHandler(gameManager *game.GameManager) *GameWSHandler {
	return &GameWSHandler{
		gameManager: gameManager,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *GameWSHandler) HostGame(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("gameId")
	hostID := r.URL.Query().Get("hostId")
	
	if gameID == "" {
		api.HandleError(w, &api.AppError{
			Code:    http.StatusBadRequest,
			Message: api.ErrMissingParam + ": gameId",
		})
		return
	}
	if hostID == "" {
		api.HandleError(w, &api.AppError{
			Code:    http.StatusBadRequest,
			Message: api.ErrMissingParam + ": hostId",
		})
		return
	}

	gameObj, err := h.gameManager.GetGame(gameID)
	if err != nil {
		api.HandleError(w, &api.AppError{
			Code:    http.StatusNotFound,
			Message: api.ErrGameNotFound,
		})
		return 
	}


	if gameObj.HostPlayerID != hostID {
		api.HandleError(w, &api.AppError{
			Code:    http.StatusUnauthorized,
			Message: api.ErrInvalidHostID,
		})
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		api.HandleError(w, &api.AppError{
			Code:    http.StatusInternalServerError,
			Message: api.ErrWebSocketUpgrade,
			Err:     err,
		})
		return
	}

	gameObj.Mutex.Lock()
	isReconnection := gameObj.HostConn != nil
	gameObj.HostConn = conn
	gameObj.LastActivity = time.Now()
	
	if isReconnection {
		for i, viewer := range gameObj.Viewers {
			if err := viewer.WriteJSON(map[string]interface{}{
				"type":    "hostReconnected",
				"message": "Host has reconnected to the game",
			}); err != nil {
				logger.Debug.Printf("Error notifying viewer %d about host reconnection: %v", i, err)
			}
		}
		logger.Info.Printf("Host reconnected: GameID=%s, HostID=%s", gameID, hostID)
	}
	gameObj.Mutex.Unlock()

	h.gameManager.Stats.Mutex.Lock()
	h.gameManager.Stats.TotalHostConnections++
	h.gameManager.Stats.Mutex.Unlock()

	logger.Info.Printf("Host connected: GameID=%s, HostID=%s", gameID, hostID)

	for {
		var message HostMessage
		if err := conn.ReadJSON(&message); err != nil {
			logger.Error.Printf("Host disconnected (GameID=%s): %v", gameID, err)
			break
		}

		

		logger.Debug.Printf("Update received: GameID=%s", gameID)

		gameObj.Mutex.Lock()
		gameObj.GameState = message.GameState
		gameObj.LastActivity = time.Now()
		viewerCount := len(gameObj.Viewers)
		
		var failedViewers []int
		for i, viewer := range gameObj.Viewers {
			if err := viewer.WriteJSON(map[string]interface{}{
				"type":      "gameState",
				"gameState": gameObj.GameState,
			}); err != nil {
				logger.Debug.Printf("Error sending to viewer %d: %v", i, err)
				failedViewers = append(failedViewers, i)
			}
		}
		
		if len(failedViewers) > 0 {
			for i := len(failedViewers) - 1; i >= 0; i-- {
				index := failedViewers[i]
				gameObj.Viewers = append(gameObj.Viewers[:index], gameObj.Viewers[index+1:]...)
				logger.Warn.Printf("Viewer removed after send failure: GameID=%s, index=%d", gameID, index)
			}
			viewerCount = len(gameObj.Viewers)
		}
		
		gameObj.Mutex.Unlock()
		
		logger.Debug.Printf("Game state broadcast to %d viewers", viewerCount)
	}

	gameObj.Mutex.Lock()
	gameObj.HostConn = nil
	gameObj.LastActivity = time.Now()
	gameObj.Mutex.Unlock()

	gameObj.Mutex.Lock()
	for i, viewer := range gameObj.Viewers {
		if err := viewer.WriteJSON(map[string]interface{}{
			"type":      "hostDisconnected",
			"message": "Host has disconnected",
		}); err != nil {
			logger.Debug.Printf("Error sending to viewer %d: %v", i, err)
		}
	}
	gameObj.Mutex.Unlock()

}

func (h *GameWSHandler) ViewGame(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("gameId")
	
	if gameID == "" {
		api.HandleError(w, &api.AppError{
			Code:    http.StatusBadRequest,
			Message: api.ErrMissingParam + ": gameId",
		})
		return
	}

	gameObj, err := h.gameManager.GetGame(gameID)
	if err != nil {
		api.HandleError(w, &api.AppError{
			Code:    http.StatusNotFound,
			Message: api.ErrGameNotFound,
		})
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		api.HandleError(w, &api.AppError{
			Code:    http.StatusInternalServerError,
			Message: api.ErrWebSocketUpgrade,
			Err:     err,
		})
		return
	}

	h.gameManager.Stats.Mutex.Lock()
	h.gameManager.Stats.TotalHostConnections++
	h.gameManager.Stats.Mutex.Unlock()

	if err := conn.WriteJSON(map[string]interface{}{
		"type":      "gameState",
		"gameState": gameObj.GameState,
	}); err != nil {
		logger.Error.Printf("Error sending initial state to viewer: %v", err)
		conn.Close()
		return
	}

	viewerCount := 0
	gameObj.Mutex.Lock()
	gameObj.LastActivity = time.Now()
	if gameObj.HostConn != nil {
		if err := gameObj.HostConn.WriteJSON(map[string]string{
			"type": "viewerJoined",
		}); err != nil {
			logger.Warn.Printf("Unable to notify host of new viewer: %v", err)
		}
	}
	
	gameObj.Viewers = append(gameObj.Viewers, conn)
	viewerCount = len(gameObj.Viewers)
	gameObj.Mutex.Unlock()

	logger.Info.Printf("New viewer connected: GameID=%s (total: %d viewers)", gameID, viewerCount)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			logger.Debug.Printf("Viewer disconnected: GameID=%s, Error: %v", gameID, err)
			break
		}
	}

	gameObj.Mutex.Lock()
	for i, v := range gameObj.Viewers {
		if v == conn {
			gameObj.Viewers = append(gameObj.Viewers[:i], gameObj.Viewers[i+1:]...)
			break
		}
	}
	viewerCount = len(gameObj.Viewers)
	gameObj.LastActivity = time.Now()
	gameObj.Mutex.Unlock()
	
	logger.Info.Printf("Viewer disconnected: GameID=%s (remaining: %d viewers)", gameID, viewerCount)
}
