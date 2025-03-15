package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/vignaliVincent/yamsAttackSocket/internal/models"
)

// ==================== TYPES & CONSTANTES ====================

// WebSocketMessage est la structure générique pour tous les messages WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`    // Type de message (action)
	GameID  string      `json:"gameId"`  // Identifiant du jeu
	Content interface{} `json:"content"` // Contenu du message (flexible)
}

// ViewerInfo contient les informations d'un spectateur
type ViewerInfo struct {
	ViewerID string `json:"viewerId"`
}

// ConnectionType définit le rôle d'une connexion WebSocket
type ConnectionType int

const (
	TypePlayer ConnectionType = iota
	TypeViewer 
)

// MessageType définit les différents types de messages supportés
const (
	TypeJoinGame       = "join_game"
	TypeUpdateGameState = "update_game_state"
	TypeLeaveGame      = "leave_game"
	TypePlayerJoined   = "player_joined"
	TypeGameStateUpdated = "game_state_updated"
	TypeViewerJoined   = "viewer_joined"
	TypeViewRequest    = "view_request"
)

// ==================== VARIABLES GLOBALES ====================

// Configuration de l'upgrader WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,  // Taille du buffer de lecture augmentée pour les états de jeu plus grands
	WriteBufferSize: 2048,  // Taille du buffer d'écriture augmentée
	CheckOrigin: func(r *http.Request) bool {
		return true // Autorise toutes les origines (pour développement)
	},
}

// Stockage des connexions WebSocket actives
var (
	// Map qui stocke les connexions par gameID
	connections = make(map[string][]*websocket.Conn)
	// Map qui stocke le type de chaque connexion (joueur ou spectateur)
	connectionTypes = make(map[*websocket.Conn]ConnectionType)
	// Mutex pour protéger l'accès concurrent aux maps
	connectionsMutex = sync.RWMutex{}
)

// ==================== GESTIONNAIRE PRINCIPAL ====================

// WebSocketHandler gère les connexions WebSocket entrantes
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier si c'est un spectateur via le paramètre de requête
	isViewer := r.URL.Query().Get("view") == "true"
	gameIDFromQuery := r.URL.Query().Get("gameId")
	
	// Upgrade la connexion HTTP vers WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Erreur lors de l'upgrade de la connexion: %v", err)
		return
	}

	// La gameID sera récupérée du premier message ou de l'URL
	var gameID string
	if isViewer && gameIDFromQuery != "" {
		gameID = gameIDFromQuery
		
		// Vérifier si le jeu partagé existe
		sharedGame, exists := GetSharedGame(gameID)
		if !exists {
			log.Printf("Tentative de connexion à un jeu inexistant: %s", gameID)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"error", "content":"Jeu introuvable"}`))
			conn.Close()
			return
		}
		
		// Enregistrer la connexion comme spectateur et lui envoyer l'état actuel
		registerViewer(conn, sharedGame)
	}
	
	// Fermeture de la connexion à la fin de la fonction
	defer cleanupConnection(conn, &gameID)
	
	log.Printf("Nouvelle connexion WebSocket établie: %s (spectateur: %v)", gameID, isViewer)
	
	// Boucle principale pour lire les messages
	messageLoop(conn, &gameID, isViewer)
}

// registerViewer enregistre une connexion comme spectateur et envoie l'état initial
func registerViewer(conn *websocket.Conn, game SharedGame) {
	// Marquer la connexion comme spectateur
	connectionsMutex.Lock()
	connectionTypes[conn] = TypeViewer
	connectionsMutex.Unlock()
	
	// Ajouter à la liste des connexions pour ce jeu
	addConnection(game.GameID, conn)
	
	// Envoyer l'état actuel du jeu au spectateur
	stateMsg := models.GameStateMessage{
		State:   game.GameState,
		Version: 1,
	}
	
	wsMsg := WebSocketMessage{
		Type:    TypeGameStateUpdated,
		GameID:  game.GameID,
		Content: stateMsg,
	}
	
	msgBytes, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Erreur lors de la sérialisation de l'état initial: %v", err)
		return
	}
	
	if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
		log.Printf("Erreur lors de l'envoi de l'état initial: %v", err)
	} else {
		log.Printf("État initial envoyé au spectateur pour le jeu %s", game.GameID)
	}
}

// messageLoop gère la lecture continue des messages d'une connexion WebSocket
func messageLoop(conn *websocket.Conn, gameID *string, isViewer bool) {
	for {
		// Lire le message du client
		_, message, err := conn.ReadMessage()
		if err != nil {
			handleReadError(err)
			break
		}
		
		// Les spectateurs ne peuvent pas envoyer de messages qui modifient l'état
		if isViewer {
			log.Printf("Message ignoré du spectateur: spectateurs en lecture seule")
			continue
		}
		
		// Décoder le message
		var wsMessage WebSocketMessage
		if err := json.Unmarshal(message, &wsMessage); err != nil {
			log.Printf("Erreur de décodage JSON: %v", err)
			continue
		}
		
		// Si gameID n'est pas encore défini (première connexion) ou change
		if *gameID == "" || *gameID != wsMessage.GameID {
			handleGameAssociation(conn, gameID, wsMessage.GameID, TypePlayer)
		}
		
		// Traiter le message selon son type
		processMessage(conn, *gameID, wsMessage)
	}
}

// ==================== GESTIONNAIRES DE MESSAGES ====================

// processMessage traite un message selon son type
func processMessage(conn *websocket.Conn, gameID string, message WebSocketMessage) {
	switch message.Type {
	case TypeJoinGame:
		handleJoinGame(gameID, conn, message)
		
	case TypeUpdateGameState:
		handleUpdateGameState(gameID, message)
		
	case TypeViewRequest:
		handleViewRequest(gameID, conn, message)
		
	case TypeLeaveGame:
		// Si le client quitte explicitement la partie
		// (traité automatiquement par le defer)
		
	default:
		log.Printf("Type de message inconnu: %s", message.Type)
	}
}

// Gère un joueur qui rejoint une partie
func handleJoinGame(gameID string, conn *websocket.Conn, message WebSocketMessage) {
	// Informer les autres joueurs qu'un nouveau joueur a rejoint
	broadcastToOthers(gameID, conn, WebSocketMessage{
		Type:    TypePlayerJoined,
		GameID:  gameID,
		Content: message.Content,
	})
	
	log.Printf("Joueur a rejoint la partie: %s", gameID)
}

// Gère une mise à jour de l'état du jeu
func handleUpdateGameState(gameID string, message WebSocketMessage) {
	contentBytes, err := json.Marshal(message.Content)
	if err != nil {
		log.Printf("Erreur lors de la conversion du contenu: %v", err)
		return
	}
	
	gameState, err := parseGameStateMessage(contentBytes)
	if err != nil {
		log.Printf("Erreur de traitement de l'état du jeu: %v", err)
		return
	}
	
	// Mettre à jour l'état dans la map des jeux partagés
	// (important pour que les nouveaux spectateurs reçoivent le dernier état)
	UpdateSharedGameState(gameID, gameState.State)
	
	// Diffuser la mise à jour à tous les clients de cette partie
	broadcastGameState(gameID, *gameState)
	log.Printf("État du jeu mis à jour et diffusé pour la partie: %s", gameID)
}

// Gère une demande de visualisation d'un jeu partagé
func handleViewRequest(gameID string, conn *websocket.Conn, message WebSocketMessage) {
	sharedGame, exists := GetSharedGame(gameID)
	if !exists {
		// Le jeu demandé n'existe pas
		errorMsg := WebSocketMessage{
			Type:   "error",
			GameID: gameID,
			Content: map[string]string{
				"message": "Jeu introuvable",
			},
		}
		
		msgBytes, _ := json.Marshal(errorMsg)
		conn.WriteMessage(websocket.TextMessage, msgBytes)
		return
	}
	
	// Marquer la connexion comme spectateur
	connectionsMutex.Lock()
	connectionTypes[conn] = TypeViewer
	connectionsMutex.Unlock()
	
	// Envoyer l'état actuel du jeu
	stateMsg := models.GameStateMessage{
		State:   sharedGame.GameState,
		Version: 1,
	}
	
	wsMsg := WebSocketMessage{
		Type:    TypeGameStateUpdated,
		GameID:  gameID,
		Content: stateMsg,
	}
	
	msgBytes, _ := json.Marshal(wsMsg)
	conn.WriteMessage(websocket.TextMessage, msgBytes)
	
	log.Printf("Demande de visualisation traitée pour le jeu: %s", gameID)
}

// ==================== GESTION DES CONNEXIONS ====================

// handleGameAssociation associe une connexion à une partie
func handleGameAssociation(conn *websocket.Conn, currentGameID *string, newGameID string, connType ConnectionType) {
	// Si c'est le premier message ou si le gameID change
	if *currentGameID == "" || *currentGameID != newGameID {
		// Si ce n'est pas le premier message, retirer d'abord de l'ancienne partie
		if *currentGameID != "" {
			removeConnection(*currentGameID, conn)
		}
		
		// Mettre à jour le gameID
		*currentGameID = newGameID
		
		// Ajouter la connexion à la nouvelle partie
		addConnection(*currentGameID, conn)
		
		// Stocker le type de connexion
		connectionsMutex.Lock()
		connectionTypes[conn] = connType
		connectionsMutex.Unlock()
	}
}

// cleanupConnection nettoie la connexion lors de la fermeture
func cleanupConnection(conn *websocket.Conn, gameID *string) {
	conn.Close()
	
	// Si une gameID a été assignée, retirer la connexion de la liste
	if *gameID != "" {
		removeConnection(*gameID, conn)
	}
	
	// Supprimer l'entrée dans la map des types de connexion
	connectionsMutex.Lock()
	delete(connectionTypes, conn)
	connectionsMutex.Unlock()
	
	log.Printf("Connexion WebSocket fermée pour la partie: %s", *gameID)
}

// handleReadError traite les erreurs de lecture WebSocket
func handleReadError(err error) {
	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		log.Printf("Erreur de lecture: %v", err)
	}
}

// Ajoute une connexion à une partie
func addConnection(gameID string, conn *websocket.Conn) {
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()
	
	connections[gameID] = append(connections[gameID], conn)
	log.Printf("Connexion ajoutée à la partie: %s (total: %d)", gameID, len(connections[gameID]))
}

// Retire une connexion d'une partie
func removeConnection(gameID string, conn *websocket.Conn) {
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()
	
	conns := connections[gameID]
	for i, c := range conns {
		if c == conn {
			// Retirer la connexion en réorganisant la slice
			connections[gameID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
	
	// Si c'était la dernière connexion pour cette partie, supprimer la clé
	if len(connections[gameID]) == 0 {
		delete(connections, gameID)
		log.Printf("Partie supprimée: %s (plus de connexions)", gameID)
	} else {
		log.Printf("Connexion retirée de la partie: %s (restantes: %d)", gameID, len(connections[gameID]))
	}
}

// ==================== MESSAGES & DIFFUSION ====================

// parseGameStateMessage analyse un message contenant l'état du jeu
func parseGameStateMessage(message []byte) (*models.GameStateMessage, error) {
	var gameStateMsg models.GameStateMessage
	err := json.Unmarshal(message, &gameStateMsg)
	if err != nil {
		return nil, err
	}
	
	return &gameStateMsg, nil
}

// broadcastGameState envoie une mise à jour de l'état du jeu à tous les joueurs d'une partie
func broadcastGameState(gameID string, gameState models.GameStateMessage) {
	// Création du message WebSocket
	wsMessage := WebSocketMessage{
		Type:    TypeGameStateUpdated,
		GameID:  gameID,
		Content: gameState,
	}
	
	// Sérialisation en JSON
	messageBytes, err := json.Marshal(wsMessage)
	if err != nil {
		log.Printf("Erreur de sérialisation du message: %v", err)
		return
	}
	
	// Envoi à tous les clients
	broadcastRawMessage(gameID, messageBytes)
}

// Diffuse un message à tous les clients d'une partie sauf l'expéditeur
func broadcastToOthers(gameID string, sender *websocket.Conn, message WebSocketMessage) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Erreur de sérialisation: %v", err)
		return
	}
	
	connectionsMutex.RLock()
	conns := connections[gameID]
	connectionsMutex.RUnlock()
	
	for _, conn := range conns {
		if conn != sender { // Ne pas envoyer à l'expéditeur
			if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
				log.Printf("Erreur d'envoi: %v", err)
			}
		}
	}
}

// broadcastRawMessage diffuse un message brut à tous les clients d'une partie
func broadcastRawMessage(gameID string, messageBytes []byte) {
	connectionsMutex.RLock()
	conns := connections[gameID]
	connectionsMutex.RUnlock()
	
	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
			log.Printf("Erreur d'envoi au client: %v", err)
		}
	}
}


