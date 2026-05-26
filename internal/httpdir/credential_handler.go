package httpdir

import (
	"encoding/json"
	"errors"
	"net/http"
	"tracelock/internal/access"

	"github.com/go-chi/chi/v5"
)

func EnrollCredentialHandler(service *access.CredentialService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req struct {
			EntryMethod    string `json:"entry_method"`
			CredentialHash string `json:"credential_hash"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		userID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid user id")
			return
		}

		credential, err := service.EnrollCredential(userID, req.EntryMethod, req.CredentialHash)
		if err != nil {
			if errors.Is(err, access.ErrCredentialExists) {
				WriteError(w, http.StatusConflict, "credential already exists")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not enroll credential")
			return
		}
		WriteJSON(w, http.StatusCreated, CredentialResponse{
			ID:             credential.ID,
			UserID:         credential.UserID,
			EntryMethod:    credential.EntryMethod,
			CredentialHash: credential.CredentialHash,
			EnrolledAt:     credential.EnrolledAt,
			Revoked:        credential.Revoked,
		})
	}
}

func GetCredentialHandler(service *access.CredentialService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		entryMethod := chi.URLParam(r, "method")

		userID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid user id")
			return
		}

		credential, err := service.GetCredential(userID, entryMethod)
		if err != nil {
			if errors.Is(err, access.ErrCredentialNotFound) {
				WriteError(w, http.StatusNotFound, "credential not found")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not fetch credential")
			return
		}
		WriteJSON(w, http.StatusOK, credential)
	}
}

func RevokeCredentialHandler(service *access.CredentialService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		entryMethod := chi.URLParam(r, "method")

		userID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid user id")
			return
		}

		if err := service.RevokeCredential(userID, entryMethod); err != nil {
			switch {
			case errors.Is(err, access.ErrCredentialNotFound):
				WriteError(w, http.StatusNotFound, "credential not found")
			default:
				WriteError(w, http.StatusInternalServerError, "could not revoke credential")
			}
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{"message": "credential revoked"})
	}
}

func ListUserCredentialsHandler(service *access.CredentialService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid user id")
			return
		}

		credentials, err := service.ListUserCredentials(userID)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "could not fetch credentials")
			return
		}
		WriteJSON(w, http.StatusOK, credentials)
	}
}
