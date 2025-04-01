package game

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type GameManagerTestSuite struct {
	suite.Suite
	Manager *GameManager
}

func (suite *GameManagerTestSuite) SetupTest() {
	suite.Manager = NewGameManager()
}

func (suite *GameManagerTestSuite) TestCreateGame() {
	t := suite.T()
	
	hostID := "player123"
	initialState := json.RawMessage(`{"state":"initial"}`)
	
	gameID, err := suite.Manager.CreateGame(hostID, initialState)
	
	assert.NoError(t, err)
	assert.NotEmpty(t, gameID)
	
	game, err := suite.Manager.GetGame(gameID)
	assert.NoError(t, err)
	assert.NotNil(t, game)
	assert.Equal(t, hostID, game.HostPlayerID)
	assert.Equal(t, initialState, game.GameState)
	
	stats := suite.Manager.GetMetrics()
	assert.Equal(t, 1, stats.ActiveGames)
	assert.Equal(t, 1, stats.TotalGamesCreated)
}

func (suite *GameManagerTestSuite) TestGetGame() {
	t := suite.T()
	
	hostID := "player456"
	initialState := []byte(`{"status": "waiting"}`)
	gameID, _ := suite.Manager.CreateGame(hostID, initialState)
	
	game, err := suite.Manager.GetGame(gameID)
	
	assert.NoError(t, err)
	assert.NotNil(t, game)
	assert.Equal(t, gameID, game.GameID)
	
	nonExistentID := uuid.New().String()
	game, err = suite.Manager.GetGame(nonExistentID)
	
	assert.Error(t, err)
	assert.Nil(t, game)
	assert.Contains(t, err.Error(), "not found")
}

func (suite *GameManagerTestSuite) TestRemoveGame() {
	t := suite.T()
	
	hostID1 := "player789"
	initialState := []byte(`{"status": "waiting"}`)
	gameID1, _ := suite.Manager.CreateGame(hostID1, initialState)
	
	hostID2 := "player987"
	gameID2, _ := suite.Manager.CreateGame(hostID2, initialState)
	
	stats := suite.Manager.GetMetrics()
	assert.Equal(t, 2, stats.ActiveGames)
	
	suite.Manager.RemoveGame(gameID1)
	
	_, err := suite.Manager.GetGame(gameID1)
	assert.Error(t, err)
	
	game, err := suite.Manager.GetGame(gameID2)
	assert.NoError(t, err)
	assert.NotNil(t, game)
	
	stats = suite.Manager.GetMetrics()
	assert.Equal(t, 1, stats.ActiveGames)
}

func (suite *GameManagerTestSuite) TestCleanupInactiveGames() {
	t := suite.T()
	
	hostID := "playerCleanup"
	initialState := []byte(`{"status": "waiting"}`)
	gameID, _ := suite.Manager.CreateGame(hostID, initialState)
	
	game, _ := suite.Manager.GetGame(gameID)
	
	game.Mutex.Lock()
	game.LastActivity = time.Now().Add(-3 * time.Hour)
	game.Mutex.Unlock()
	
	suite.Manager.CleanupInactiveGames()
	
	_, err := suite.Manager.GetGame(gameID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	
	stats := suite.Manager.GetMetrics()
	assert.Equal(t, 0, stats.ActiveGames)
}

func (suite *GameManagerTestSuite) TestUpdateViewerCount() {
	t := suite.T()
	
	initialStats := suite.Manager.GetMetrics()
	
	suite.Manager.UpdateViewerCount(3)
	
	stats := suite.Manager.GetMetrics()
	assert.Equal(t, initialStats.TotalViewers+3, stats.TotalViewers)
	
	suite.Manager.UpdateViewerCount(-1)
	
	stats = suite.Manager.GetMetrics()
	assert.Equal(t, initialStats.TotalViewers+2, stats.TotalViewers)
}

func (suite *GameManagerTestSuite) TestUpdateHostCount() {
	t := suite.Suite.T()
	
	initialStats := suite.Manager.GetMetrics()
	
	suite.Manager.UpdateHostCount(2)
	
	stats := suite.Manager.GetMetrics()
	assert.Equal(t, initialStats.TotalHostConnections+2, stats.TotalHostConnections)
	
	suite.Manager.UpdateHostCount(-1)
	
	stats = suite.Manager.GetMetrics()
	assert.Equal(t, initialStats.TotalHostConnections+1, stats.TotalHostConnections)
}

func (suite *GameManagerTestSuite) TestGetMetrics() {
	t := suite.T()
	
	suite.Manager.CreateGame("player1", []byte(`{}`))
	suite.Manager.CreateGame("player2", []byte(`{}`))
	
	suite.Manager.UpdateViewerCount(5)
	suite.Manager.UpdateHostCount(2)
	
	stats := suite.Manager.GetMetrics()
	
	assert.Equal(t, 2, stats.ActiveGames)
	assert.Equal(t, 2, stats.TotalGamesCreated)
	assert.Equal(t, 5, stats.TotalViewers)
	assert.Equal(t, 2, stats.TotalHostConnections)
	assert.False(t, stats.StartTime.IsZero())
}

func (suite *GameManagerTestSuite) TestCleanupInactiveGamesInterval() {
	t := suite.T()
	
	activeGameID, _ := suite.Manager.CreateGame("activePlayer", []byte(`{}`))
	inactiveGameID, _ := suite.Manager.CreateGame("inactivePlayer", []byte(`{}`))
	
	inactiveGame, _ := suite.Manager.GetGame(inactiveGameID)
	inactiveGame.Mutex.Lock()
	inactiveGame.LastActivity = time.Now().Add(-3 * time.Hour)
	inactiveGame.Mutex.Unlock()
	
	suite.Manager.CleanupInactiveGames()
	
	_, err := suite.Manager.GetGame(inactiveGameID)
	assert.Error(t, err)
	
	activeGame, err := suite.Manager.GetGame(activeGameID)
	assert.NoError(t, err)
	assert.NotNil(t, activeGame)
}

func TestGameManagerSuite(t *testing.T) {
	suite.Run(t, new(GameManagerTestSuite))
}