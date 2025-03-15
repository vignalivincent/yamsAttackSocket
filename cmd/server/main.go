package main

import (
	"log"
	"net/http"
	"os"

	"github.com/vignaliVincent/yamsAttackSocket/internal/handlers"
	"github.com/vignaliVincent/yamsAttackSocket/internal/middleware"
)

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    // Créer les handlers avec CORS
    wsHandler := http.HandlerFunc(handlers.WebSocketHandler)
    shareHandler := http.HandlerFunc(handlers.ShareHandler)
    
    // Enregistrer les routes avec middleware CORS
    http.Handle("/ws", middleware.CorsMiddleware(wsHandler))
    http.Handle("/share", middleware.CorsMiddleware(shareHandler))
    
    log.Printf("Serveur démarré sur le port %s", port)
    
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatal("Erreur lors du démarrage du serveur:", err)
    }
}
