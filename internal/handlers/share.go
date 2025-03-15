package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vignaliVincent/yamsAttackSocket/internal/models"
	"github.com/vignaliVincent/yamsAttackSocket/internal/services"
)

// Structure pour stocker les informations sur les jeux partagés
type SharedGame struct {
	GameID        string              // ID unique du jeu
	PrimaryPlayer string              // ID du joueur principal qui partage le jeu
	GameState     models.GameState    // État actuel du jeu
	CreatedAt     time.Time           // Date de création
	ViewerPhones  []string            // Numéros de téléphone des spectateurs
}

// Map pour stocker les jeux partagés, accessible par GameID
var sharedGames = make(map[string]SharedGame)
var sharedGamesMutex = sync.RWMutex{}

// ShareRequest représente la structure de la requête pour partager un jeu
type ShareRequest struct {
	PlayerID     string        `json:"playerId"`      // ID du joueur principal
	GameData     models.GameState `json:"gameData"`   // Données initiales du jeu
	PhoneNumbers []string      `json:"phoneNumbers"`  // Numéros de téléphone des spectateurs
}

// ShareResponse représente la structure de la réponse envoyée au client
type ShareResponse struct {
	Success bool   `json:"success"`
	GameID  string `json:"gameId"`
	ViewURL string `json:"viewUrl"`
	Error   string `json:"error,omitempty"`
}

// ShareHandler gère les requêtes pour partager un jeu avec des spectateurs
func ShareHandler(w http.ResponseWriter, r *http.Request) {
	// Ajouter les en-têtes CORS directement dans le handler
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	
	// Gestion des requêtes OPTIONS (preflight)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Vérifier que la méthode est POST
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Décoder le corps de la requête
	var shareReq ShareRequest
	err := json.NewDecoder(r.Body).Decode(&shareReq)
	if err != nil {
		http.Error(w, "Format de requête invalide: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Valider la requête
	if shareReq.PlayerID == "" {
		http.Error(w, "PlayerID est requis", http.StatusBadRequest)
		return
	}
	if len(shareReq.PhoneNumbers) == 0 {
		http.Error(w, "Au moins un numéro de téléphone est requis", http.StatusBadRequest)
		return
	}

	// Log de la requête
	log.Printf("Requête de partage reçue pour le joueur %s avec %d numéros", 
		shareReq.PlayerID, len(shareReq.PhoneNumbers))
		
	// Générer un identifiant unique pour le jeu
	gameID := uuid.New().String()
	
	// Créer un nouveau jeu partagé
	sharedGame := SharedGame{
		GameID:        gameID,
		PrimaryPlayer: shareReq.PlayerID,
		GameState:     shareReq.GameData,
		CreatedAt:     time.Now(),
		ViewerPhones:  shareReq.PhoneNumbers,
	}
	
	// Stocker le jeu dans la map des jeux partagés (avec protection mutex)
	sharedGamesMutex.Lock()
	sharedGames[gameID] = sharedGame
	sharedGamesMutex.Unlock()
	
	// Construire l'URL de visualisation (utiliser l'origine de la requête)
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "http://" + r.Host
	}
	
	viewURL, err := url.Parse(origin)
	if err != nil {
		log.Printf("Erreur lors de l'analyse de l'URL d'origine: %v", err)
		viewURL, _ = url.Parse("http://localhost:3000") // URL par défaut
	}
	
	// Ajouter les paramètres pour la vue
	q := viewURL.Query()
	q.Add("view", "true")
	q.Add("gameId", gameID)
	viewURL.RawQuery = q.Encode()
	
	// Préparer le message SMS
	smsMessage := fmt.Sprintf(
		"Vous avez été invité à suivre une partie de Yams Attack! Cliquez sur ce lien pour voir la partie en direct: %s",
		viewURL.String(),
	)
	
	// Envoyer les SMS aux spectateurs
	success, failures := services.DefaultSMSService.SendBulkSMS(shareReq.PhoneNumbers, smsMessage)
	
	log.Printf(
		"SMS envoyés pour le jeu %s: %d réussis, %d échecs",
		gameID,
		len(success),
		len(failures),
	)
	
	if len(failures) > 0 {
		log.Printf("Erreurs d'envoi SMS: %v", failures)
	}
	
	// Préparer la réponse
	resp := ShareResponse{
		Success: true,
		GameID:  gameID,
		ViewURL: viewURL.String(),
	}

	// Envoyer la réponse
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetSharedGame récupère un jeu partagé par son ID
func GetSharedGame(gameID string) (SharedGame, bool) {
	sharedGamesMutex.RLock()
	defer sharedGamesMutex.RUnlock()
	
	game, exists := sharedGames[gameID]
	return game, exists
}

// UpdateSharedGameState met à jour l'état d'un jeu partagé
func UpdateSharedGameState(gameID string, newState models.GameState) bool {
	sharedGamesMutex.Lock()
	defer sharedGamesMutex.Unlock()
	
	if game, exists := sharedGames[gameID]; exists {
		game.GameState = newState
		sharedGames[gameID] = game
		return true
	}
	return false
}
