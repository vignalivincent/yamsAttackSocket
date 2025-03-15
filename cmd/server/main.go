package main

import (
	"log"
	"net/http"
	"os"

	"github.com/vignaliVincent/yamsAttackSocket/internal/handlers"
)

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    // Initialiser les routes
    http.HandleFunc("/ws", handlers.WebSocketHandler)
    http.HandleFunc("/share", handlers.ShareHandler)
    
    // Servir les fichiers statiques (pour le front-end)
    http.Handle("/", http.FileServer(http.Dir("./static")))
    
    log.Printf("Serveur démarré sur le port %s", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatal("Erreur lors du démarrage du serveur:", err)
    }
}