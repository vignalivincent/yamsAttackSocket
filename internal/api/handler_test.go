package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vincentvignali/yamsAttackSocket/internal/game"
)

func TestInitSharedGame_Integration(t *testing.T) {
	gameManager := game.NewGameManager()
	handler := NewGameHTTPHandler(gameManager)

	t.Run("Successful Game Creation", func(t *testing.T) {
		reqBody := InitGameRequest{
			HostPlayerID: "player1", 
			GameState:    []byte(`{"state": "initial"}`),
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/init-game", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Origin", "http://example.com")

		w := httptest.NewRecorder()
		handler.InitSharedGame(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status OK")

		var response InitGameResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Should parse response successfully")
		
		assert.NotEmpty(t, response.GameID, "GameID should not be empty")
		assert.NotEmpty(t, response.ShareURL, "ShareURL should not be empty")

		game, err := gameManager.GetGame(response.GameID)
		assert.NoError(t, err, "Game should exist in GameManager")
		assert.NotNil(t, game, "Game should not be nil")
	})

	t.Run("Invalid HTTP Method", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/init-game", nil)
		w := httptest.NewRecorder()
		handler.InitSharedGame(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code, "Expected method not allowed")
	})

	t.Run("Empty Request Body", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/init-game", nil)
		w := httptest.NewRecorder()
		handler.InitSharedGame(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected bad request")
	})
}

func TestServerStats_Integration(t *testing.T) {
	gameManager := game.NewGameManager()
	handler := NewGameHTTPHandler(gameManager)

	game1ID, _ := gameManager.CreateGame("player1", []byte(`{"state": "game1"}`))
	game2ID, _ := gameManager.CreateGame("player2", []byte(`{"state": "game2"}`))
	
	t.Logf("Created test games with IDs: %s, %s", game1ID, game2ID)

	req, _ := http.NewRequest(http.MethodGet, "/server-stats", nil)
	w := httptest.NewRecorder()
	handler.ServerStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status OK")

	var metrics game.ServerStatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &metrics)
	assert.NoError(t, err, "Should parse metrics successfully")
	
	assert.GreaterOrEqual(t, metrics.TotalGamesCreated, int(2), "Should have created at least 2 games")
	assert.NotEmpty(t, metrics.Uptime, "Uptime should not be empty")
	
	assert.GreaterOrEqual(t, metrics.ActiveGames, int(2), "Should have at least 2 active games")
	
	t.Logf("Server metrics: %+v", metrics)
}