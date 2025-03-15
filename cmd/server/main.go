package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

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
    
    // Servir les fichiers de test statiques
    fs := http.FileServer(http.Dir(filepath.Join(getProjectRoot(), "test")))
    http.Handle("/test/", http.StripPrefix("/test/", middleware.CorsMiddleware(fs)))
    
    log.Printf("Serveur démarré sur le port %s", port)
    log.Printf("Pages de test disponibles à: http://localhost:%s/test/player_test.html", port)
    
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatal("Erreur lors du démarrage du serveur:", err)
    }
}

// getProjectRoot tente de déterminer le chemin racine du projet
func getProjectRoot() string {
    // Par défaut, utiliser le répertoire courant
    dir, err := os.Getwd()
    if err != nil {
        return "."
    }
    
    // Si nous sommes dans cmd/server, remonter de deux niveaux
    if filepath.Base(dir) == "server" && filepath.Base(filepath.Dir(dir)) == "cmd" {
        return filepath.Join(dir, "../..")
    }
    
    return dir
}