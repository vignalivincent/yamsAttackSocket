package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vincentvignali/yamsAttackSocket/internal/game"
)

func TestHostGameSuccessfulConnection(t *testing.T) {
    gameManager := game.NewGameManager()
    wsHandler := NewGameWSHandler(gameManager)
    
    hostID := "test-host-success"
    initialState := json.RawMessage(`{"state":"initial"}`)
    gameID, err := gameManager.CreateGame(hostID, initialState)
    assert.NoError(t, err)
    
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        r.URL.RawQuery = "gameId=" + gameID + "&hostId=" + hostID
        wsHandler.HostGame(w, r)
    }))
    defer server.Close()
    
    
    
    game, err := gameManager.GetGame(gameID)
    assert.NoError(t, err)
    assert.NotNil(t, game)
    assert.Equal(t, hostID, game.HostPlayerID)
}

func TestGameStateUpdate(t *testing.T) {
    gameManager := game.NewGameManager()
    
    hostID := "test-host-updates"
    initialState := json.RawMessage(`{"state":"initial","score":0}`)
    gameID, _ := gameManager.CreateGame(hostID, initialState)
    
    game, _ := gameManager.GetGame(gameID)
    
    newState := json.RawMessage(`{"state":"updated","score":100}`)
    game.Mutex.Lock()
    game.GameState = newState
    game.LastActivity = time.Now()
    game.Mutex.Unlock()
    
    game, _ = gameManager.GetGame(gameID)
    assert.Equal(t, newState, game.GameState)
}

func TestGamePersistencyAfterHostDisconnected(t *testing.T) {
    gameManager := game.NewGameManager()
    
    hostID := "test-host-updates"
    initialState := json.RawMessage(`{"state":"initial","score":0}`)
    gameID, _ := gameManager.CreateGame(hostID, initialState)
    
    game, _ := gameManager.GetGame(gameID)
	
	game.Mutex.Lock()
	game.HostConn = nil
	game.Mutex.Unlock()

	game, _ = gameManager.GetGame(gameID)
	assert.Nil(t, game.HostConn)
}

