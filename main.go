package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type GameState struct {
	Players []Player `json:"players"`
	Round   int      `json:"round"`
}

type Player struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

var gameState = GameState{
	Players: []Player{},
	Round:   1,
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/game", gameHandler)
	http.HandleFunc("/player", playerHandler)
	
	// Prioritize PORT from environment variables (for Fly.io and other cloud platforms)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}
	
	fmt.Printf("Server starting on port %s...\n", port)
	log.Printf("Available endpoints:\n - GET /health\n - GET /game\n - POST /player\n")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /health", r.Method)
	
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	log.Printf("Health check: OK")
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /game", r.Method)
	w.Header().Set("Content-Type", "application/json")
	
	switch r.Method {
	case http.MethodGet:
		// Return current game state
		json.NewEncoder(w).Encode(gameState)
		log.Printf("Game state returned: %d players, round %d", len(gameState.Players), gameState.Round)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Printf("Method not allowed: %s", r.Method)
	}
}

func playerHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /player", r.Method)
	w.Header().Set("Content-Type", "application/json")
	
	switch r.Method {
	case http.MethodPost:
		var player Player
		if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Printf("Error decoding player data: %v", err)
			return
		}
		
		// Add player to game
		gameState.Players = append(gameState.Players, player)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(player)
		log.Printf("Player added: %s (ID: %s)", player.Name, player.ID)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Printf("Method not allowed: %s", r.Method)
	}
}
