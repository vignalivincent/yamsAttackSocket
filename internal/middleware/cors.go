package middleware

import "net/http"

// CorsMiddleware ajoute les en-têtes CORS nécessaires pour permettre les requêtes cross-origin
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ajouter les en-têtes CORS
		w.Header().Set("Access-Control-Allow-Origin", "*") // Permet toutes les origines
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		// Gestion des requêtes OPTIONS (preflight)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Passer au handler suivant
		next.ServeHTTP(w, r)
	})
}
