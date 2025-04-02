package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/vincentvignali/yamsAttackSocket/internal/game"
)

// WebSocketTestSuite définit une suite de tests pour les handlers WebSocket
type WebSocketTestSuite struct {
	suite.Suite
	GameManager  *game.GameManager
	WSHandler    *GameWSHandler
	Server       *httptest.Server
	HostID       string
	GameID       string
	InitialState json.RawMessage
}

// SetupTest initialise l'environnement de test avant chaque test
func (suite *WebSocketTestSuite) SetupTest() {
	suite.GameManager = game.NewGameManager()
	suite.WSHandler = NewGameWSHandler(suite.GameManager)
	suite.HostID = "test-host-id"
	suite.InitialState = json.RawMessage(`{"state":"initial","score":0}`)

	// Créer une partie
	gameID, err := suite.GameManager.CreateGame(suite.HostID, suite.InitialState)
	assert.NoError(suite.T(), err)
	suite.GameID = gameID

	// Configurer un serveur de test
	suite.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Le handler sera défini dans chaque test
	}))
}

// TearDownTest nettoie après chaque test
func (suite *WebSocketTestSuite) TearDownTest() {
	if suite.Server != nil {
		suite.Server.Close()
	}
}

// ConnectHost établit une connexion WebSocket en tant qu'hôte
func (suite *WebSocketTestSuite) ConnectHost() (*websocket.Conn, *http.Response, error) {
	wsURL := "ws" + strings.TrimPrefix(suite.Server.URL, "http") +
		"?gameId=" + suite.GameID + "&hostId=" + suite.HostID

	dialer := websocket.Dialer{}
	return dialer.Dial(wsURL, nil)
}

// TestHostGameSuccessfulConnection vérifie qu'un hôte peut se connecter à une partie
func (suite *WebSocketTestSuite) TestHostGameSuccessfulConnection() {
	t := suite.T()

	// Configurer le handler pour ce test
	suite.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.RawQuery = "gameId=" + suite.GameID + "&hostId=" + suite.HostID
		suite.WSHandler.HostGame(w, r)
	}))

	// Établir une connexion WebSocket
	conn, _, err := suite.ConnectHost()
	if err != nil {
		t.Fatalf("Could not connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Attendre que la connexion soit traitée
	time.Sleep(100 * time.Millisecond)

	// Vérifier l'état de la partie
	gameInstance, err := suite.GameManager.GetGame(suite.GameID)
	assert.NoError(t, err)
	assert.NotNil(t, gameInstance)
	assert.Equal(t, suite.HostID, gameInstance.HostPlayerID)
	assert.Equal(t, game.HostConnected, gameInstance.HostConnectionState)
}

// TestGameStateUpdate vérifie qu'une mise à jour de l'état de la partie est reflétée
func (suite *WebSocketTestSuite) TestGameStateUpdate() {
	t := suite.T()

	// Configurer le handler pour ce test
	suite.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.RawQuery = "gameId=" + suite.GameID + "&hostId=" + suite.HostID
		suite.WSHandler.HostGame(w, r)
	}))

	// Établir une connexion WebSocket
	conn, _, err := suite.ConnectHost()
	if err != nil {
		t.Fatalf("Could not connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Attendre que la connexion soit traitée
	time.Sleep(100 * time.Millisecond)

	// Mettre à jour l'état de la partie
	gameInstance, _ := suite.GameManager.GetGame(suite.GameID)
	newState := json.RawMessage(`{"state":"updated","score":100}`)

	gameInstance.Mutex.Lock()
	gameInstance.GameState = newState
	gameInstance.LastActivity = time.Now()
	gameInstance.Mutex.Unlock()

	// Vérifier que l'état a été mis à jour
	gameInstance, _ = suite.GameManager.GetGame(suite.GameID)
	assert.Equal(t, newState, gameInstance.GameState)
}

// TestGamePersistencyAfterHostDisconnected vérifie que la partie persiste après déconnexion de l'hôte
func (suite *WebSocketTestSuite) TestGamePersistencyAfterHostDisconnected() {
	t := suite.T()

	// Configurer le handler pour ce test
	suite.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.RawQuery = "gameId=" + suite.GameID + "&hostId=" + suite.HostID
		suite.WSHandler.HostGame(w, r)
	}))

	// Vérifier l'état initial
	gameInstance, _ := suite.GameManager.GetGame(suite.GameID)
	assert.Equal(t, game.HostNeverConnected, gameInstance.HostConnectionState)

	// Établir une connexion WebSocket
	conn, _, err := suite.ConnectHost()
	if err != nil {
		t.Fatalf("Could not connect to WebSocket: %v", err)
	}

	// Attendre que la connexion soit traitée
	time.Sleep(100 * time.Millisecond)

	// Vérifier que l'hôte est connecté
	gameInstance, _ = suite.GameManager.GetGame(suite.GameID)
	assert.Equal(t, game.HostConnected, gameInstance.HostConnectionState)

	// Fermer la connexion
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	conn.Close()

	// Attendre que la déconnexion soit traitée
	time.Sleep(100 * time.Millisecond)

	// Vérifier que l'hôte est déconnecté mais que la partie persiste
	gameInstance, _ = suite.GameManager.GetGame(suite.GameID)
	assert.Equal(t, game.HostDisconnected, gameInstance.HostConnectionState)
	assert.NotNil(t, gameInstance)
	assert.Equal(t, suite.GameID, gameInstance.GameID)
}

// TestHostConnectionStateTransitions vérifie les transitions d'état de connexion de l'hôte
func (suite *WebSocketTestSuite) TestHostConnectionStateTransitions() {
	t := suite.T()

	// Configurer le handler pour ce test
	suite.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.RawQuery = "gameId=" + suite.GameID + "&hostId=" + suite.HostID
		suite.WSHandler.HostGame(w, r)
	}))

	// Vérifier l'état initial
	gameInstance, _ := suite.GameManager.GetGame(suite.GameID)
	assert.Equal(t, game.HostNeverConnected, gameInstance.HostConnectionState)

	// Canal pour coordonner la fermeture du test
	done := make(chan struct{})
	defer close(done)

	// Connexion de l'hôte dans une goroutine
	go func() {
		dialer := websocket.Dialer{}
		wsURL := "ws" + strings.TrimPrefix(suite.Server.URL, "http") +
			"?gameId=" + suite.GameID + "&hostId=" + suite.HostID

		conn, _, err := dialer.Dial(wsURL, nil)
		if err != nil {
			t.Logf("Error connecting to WebSocket: %v", err)
			return
		}
		defer conn.Close()

		// Maintenir la connexion pendant un court moment
		select {
		case <-done:
			return
		case <-time.After(100 * time.Millisecond):
			// Laisser le temps au serveur de traiter la connexion
		}

		// Fermer explicitement la connexion pour tester la déconnexion
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}()

	// Attendre que la connexion soit établie
	time.Sleep(100 * time.Millisecond)

	// Vérifier que l'hôte est connecté
	gameInstance, _ = suite.GameManager.GetGame(suite.GameID)
	assert.Equal(t, game.HostConnected, gameInstance.HostConnectionState)

	// Attendre que la déconnexion soit traitée
	time.Sleep(200 * time.Millisecond)

	// Vérifier que l'hôte est déconnecté
	gameInstance, _ = suite.GameManager.GetGame(suite.GameID)
	assert.Equal(t, game.HostDisconnected, gameInstance.HostConnectionState)
}

// TestWebSocketSuite lance la suite de tests
func TestWebSocketSuite(t *testing.T) {
	suite.Run(t, new(WebSocketTestSuite))
}
