// add JSON request validation and standardized error responses
package http

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, ErrorResponse{
		Error: message,
	})
}
