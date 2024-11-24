// internal/handlers/handlers.go

package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// Handlers holds our handler dependencies
type Handlers struct {
	db *sql.DB
}

// NewHandlers creates a new Handlers instance
func NewHandlers(db *sql.DB) *Handlers {
	return &Handlers{db: db}
}

// envelope is a generic response wrapper
type envelope map[string]interface{}

// respondWithJSON sends a JSON response with proper headers
func (h *Handlers) respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	// Set content type before writing response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		// Wrap the response data in an envelope
		env := envelope{"data": data}
		err := json.NewEncoder(w).Encode(env)
		if err != nil {
			// If JSON encoding fails, log it and send a 500
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// respondWithError sends an error response with proper headers
func (h *Handlers) respondWithError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	env := envelope{"error": message}
	json.NewEncoder(w).Encode(env)
}
