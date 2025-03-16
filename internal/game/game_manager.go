package game

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/vincentvignali/yamsAttackSocket/internal/logger"
)

/*
GameManager Implementation

This file implements the GameManager, which is the central component responsible for:
- Creating, tracking, and managing game instances
- Handling game lifecycle (creation, access, removal)
- Managing websocket connections for hosts and viewers
- Performing periodic cleanup of inactive games
- Collecting and providing server metrics and statistics

The GameManager acts as a service layer between the API handlers (HTTP and WebSocket)
and the actual game instances. It ensures thread-safe access to game data through
mutex-protected operations and provides methods for updating game state.

Key responsibilities:
- Game instance creation with unique IDs
- Game lookup and retrieval
- Game cleanup and resource management
- Statistics tracking for monitoring
*/



func NewGameManager() *GameManager {
	manager := &GameManager{
		games: make(map[string]*Game),
		Stats: &ServerStats{
			StartTime: time.Now(),
		},
	}
	
	go manager.cleanupInactiveGames()
	
	return manager
}

func (m *GameManager) CreateGame(hostPlayerID string, initialState []byte) (string, error) {
	gameID := uuid.New().String()
	now := time.Now()
	
	game := &Game{
		GameID:       gameID,
		HostPlayerID: hostPlayerID,
		GameState:    initialState,
		Viewers:      make([]*websocket.Conn, 0),
		CreatedAt:    now,
		LastActivity: now,
	}
	
	m.gamesMutex.Lock()
	m.games[gameID] = game
	gameCount := len(m.games)
	m.gamesMutex.Unlock()
	
	m.Stats.Mutex.Lock()
	m.Stats.ActiveGames = gameCount
	m.Stats.TotalGamesCreated++
	m.Stats.Mutex.Unlock()
	
	logger.Info.Printf("New game created: ID=%s, Host=%s (Total: %d active games)", gameID, hostPlayerID, gameCount)
	
	return gameID, nil
}

func (m *GameManager) GetGame(gameID string) (*Game, bool) {
	m.gamesMutex.Lock()
	defer m.gamesMutex.Unlock()
	game, exists := m.games[gameID]
	return game, exists
}

func (m *GameManager) RemoveGame(gameID string) {
	m.gamesMutex.Lock()
	delete(m.games, gameID)
	gameCount := len(m.games)
	m.gamesMutex.Unlock()
	
	m.Stats.Mutex.Lock()
	m.Stats.ActiveGames = gameCount
	m.Stats.Mutex.Unlock()
	
	logger.Info.Printf("Game removed: GameID=%s (Remaining: %d active games)", gameID, gameCount)
}

func (m *GameManager) cleanupInactiveGames() {
	for {
		time.Sleep(5 * time.Minute)
		now := time.Now()
		var toDelete []string
		
		m.gamesMutex.Lock()
		for id, game := range m.games {
			game.Mutex.Lock()
			inactiveTime := now.Sub(game.LastActivity)
			creationTime := now.Sub(game.CreatedAt)
			game.Mutex.Unlock()
			
			if inactiveTime > 2*time.Hour || creationTime > 12*time.Hour {
				toDelete = append(toDelete, id)
			}
		}
		
		for _, id := range toDelete {
			game := m.games[id]
			
			game.Mutex.Lock()
			if game.HostConn != nil {
				game.HostConn.Close()
			}
			for _, viewer := range game.Viewers {
				viewer.Close()
			}
			game.Mutex.Unlock()
			
			delete(m.games, id)
			logger.Warn.Printf("Cleanup: game removed due to inactivity: GameID=%s (inactive for %v)", id, now.Sub(game.LastActivity))
		}
		
		gameCount := len(m.games)
		m.gamesMutex.Unlock()
		
		m.Stats.Mutex.Lock()
		m.Stats.ActiveGames = gameCount
		m.Stats.Mutex.Unlock()
		
		logger.System.Printf("Cleanup completed: %d games removed, %d active games remaining", len(toDelete), gameCount)
	}
}

func (m *GameManager) GetMetrics() ServerStats {
	m.Stats.Mutex.RLock()
	defer m.Stats.Mutex.RUnlock()
	
	return ServerStats{
		TotalGamesCreated:    m.Stats.TotalGamesCreated,
		ActiveGames:          m.Stats.ActiveGames,
		TotalViewers:         m.Stats.TotalViewers,
		TotalHostConnections: m.Stats.TotalHostConnections,
		StartTime:            m.Stats.StartTime,
	}
}

func (m *GameManager) UpdateViewerCount(delta int) {
	m.Stats.Mutex.Lock()
	defer m.Stats.Mutex.Unlock()
	
	m.Stats.TotalViewers += delta
}

func (m *GameManager) UpdateHostCount(delta int) {
	m.Stats.Mutex.Lock()
	defer m.Stats.Mutex.Unlock()
	
	m.Stats.TotalHostConnections += delta
}
