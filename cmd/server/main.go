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
	
	mux.HandleFunc("/initSharedGame", api.WithMiddlewares(gameHandler.InitSharedGame, api.WithCORS, api.WithLogging))
	mux.HandleFunc("/stats", api.WithMiddlewares(gameHandler.ServerStats, api.WithCORS, api.WithLogging))
	mux.HandleFunc("/hostGame", api.WithMiddlewares(wsHandler.HostGame, api.WithCORS, api.WithLogging))
	mux.HandleFunc("/viewGame", api.WithMiddlewares(wsHandler.ViewGame, api.WithCORS, api.WithLogging))
	
	
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