package main

import (
	"net/http"
	"os"
	"time"

	"github.com/vincentvignali/yamsAttackSocket/internal/api"
	"github.com/vincentvignali/yamsAttackSocket/internal/game"
	"github.com/vincentvignali/yamsAttackSocket/internal/logger"
	"github.com/vincentvignali/yamsAttackSocket/internal/websocket"
)

func main() {
	gameManager := game.NewGameManager()
	
	gameHandler := api.NewGameHTTPHandler(gameManager)
	wsHandler := websocket.NewGameWSHandler(gameManager)
	
	mux := http.NewServeMux()
	mux.HandleFunc("/initSharedGame", api.WithLogging(gameHandler.InitSharedGame))
	mux.HandleFunc("/stats", api.WithLogging(gameHandler.ServerStats))
	mux.HandleFunc("/hostGame", api.WithLogging(wsHandler.HostGame))
	mux.HandleFunc("/viewGame", api.WithLogging(wsHandler.ViewGame))
	
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	
	logger.System.Printf("Server started on port :%s", port)
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	err := server.ListenAndServe()
	if err != nil {
		logger.Error.Printf("Cannot start server: %v", err)
	}
}