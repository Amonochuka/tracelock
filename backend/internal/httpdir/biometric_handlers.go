package httpdir

import (
	"encoding/json"
	"errors"
	"net/http"

	"tracelock/internal/access"
)

func AuthenticateBiometricHandler(service *access.BiometricService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			DeviceID       int    `json:"device_id"`
			CredentialHash string `json:"credential_hash"`
			Action         string `json:"action"` // "enter" or "exit"
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.DeviceID <= 0 {
			WriteError(w, http.StatusBadRequest, "invalid device_id")
			return
		}

		if req.CredentialHash == "" {
			WriteError(w, http.StatusBadRequest, "credential_hash is required")
			return
		}

		if req.Action == "" {
			req.Action = "enter"
		}

		token, err := service.AuthenticateBiometric(req.DeviceID, req.CredentialHash, req.Action)
		if err != nil {
			switch {
			case errors.Is(err, access.ErrDeviceNotFound):
				WriteError(w, http.StatusNotFound, "device not found")
			case errors.Is(err, access.ErrDeviceInactive):
				WriteError(w, http.StatusForbidden, "device is not active")
			case errors.Is(err, access.ErrCredentialNotFound):
				WriteError(w, http.StatusNotFound, "credential not found")
			case errors.Is(err, access.ErrCredentialRevoked):
				WriteError(w, http.StatusForbidden, "credential has been revoked")
			case errors.Is(err, access.ErrAccessDenied):
				WriteError(w, http.StatusForbidden, "access denied")
			case errors.Is(err, access.ErrZoneFull):
				WriteError(w, http.StatusForbidden, "zone is full")
			case errors.Is(err, access.ErrUserAlreadyInZone):
				WriteError(w, http.StatusConflict, "user already in zone")
			default:
				WriteError(w, http.StatusInternalServerError, "authentication failed")
			}
			return
		}

		WriteJSON(w, http.StatusOK, map[string]any{
			"message": "access granted",
			"token":   token,
		})
	}
}
