package api

import (
	"fmt"
	"net/http"

	"github.com/vincentvignali/yamsAttackSocket/internal/logger"
)

const (
	ErrMethodNotAllowed = "Method not allowed"
	ErrJSONParsing      = "JSON parsing error"
	ErrGameNotFound     = "Game not found"
	ErrInvalidHostID    = "Invalid host ID"
	ErrWebSocketUpgrade = "WebSocket upgrade error"
	ErrNoBody = "Request body is empty or missing"
	ErrMissingParam     = "Required parameter missing"
)

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func HandleError(w http.ResponseWriter, appErr *AppError) {
	logger.Error.Printf("%s (Code: %d)", appErr.Error(), appErr.Code)
	http.Error(w, appErr.Message, appErr.Code)
}