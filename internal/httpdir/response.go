// add JSON request validation and standardized error responses
package httpdir

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		log.Printf("failed to encode JSON response: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, ErrorResponse{
		Error: message,
	})
}

type UserResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type ZoneResponse struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	MaxCapacity      int       `json:"max_capacity"`
	RequiresExitScan bool      `json:"requires_exit_scan"`
	CreatedAt        time.Time `json:"created_at"`
}

type DeviceResponse struct {
	ID        int       `json:"id"`
	ZoneID    int       `json:"zone_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Serial    string    `json:"serial"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

type CredentialResponse struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	EntryMethod    string    `json:"entry_method"`
	CredentialHash string    `json:"credential_hash"`
	EnrolledAt     time.Time `json:"enrolled_at"`
	Revoked        bool      `json:"revoked"`
}
