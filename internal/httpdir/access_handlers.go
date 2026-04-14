package httpdir

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"tracelock/internal/access"
	"tracelock/internal/auth"
)

func EnterZoneHandler(service *access.ZoneService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		userID, err := auth.GetUserIDFromContext(r.Context())
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		role, err := auth.GetUserRoleFromContext(r.Context())
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req struct {
			ZoneID int `json:"zone_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.ZoneID <= 0 {
			WriteError(w, http.StatusBadRequest, "invalid zone_id")
			return
		}

		err = service.HandleZoneEvent(userID, req.ZoneID, role, "enter", time.Now())
		if err != nil {
			switch {
			case errors.Is(err, access.ErrAccessDenied):
				WriteError(w, http.StatusForbidden, "you do not have access to this zone")
			case errors.Is(err, access.ErrZoneFull):
				WriteError(w, http.StatusForbidden, "zone is full")
			case errors.Is(err, access.ErrUserAlreadyInZone):
				WriteError(w, http.StatusConflict, "user already in zone")
			case errors.Is(err, access.ErrZoneNotFound):
				WriteError(w, http.StatusNotFound, "zone not found")
			default:
				WriteError(w, http.StatusInternalServerError, "internal server error")
			}
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{
			"message": "entered zone successfully",
		})
	}
}

func ExitZoneHandler(service *access.ZoneService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		userID, err := auth.GetUserIDFromContext(r.Context())
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		role, err := auth.GetUserRoleFromContext(r.Context())
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req struct {
			ZoneID int `json:"zone_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.ZoneID <= 0 {
			WriteError(w, http.StatusBadRequest, "invalid zone_id")
			return
		}

		err = service.HandleZoneEvent(userID, req.ZoneID, role, "exit", time.Now())
		if err != nil {
			switch {
			case errors.Is(err, access.ErrNoActiveSession):
				WriteError(w, http.StatusNotFound, "no active session found")
			case errors.Is(err, access.ErrZoneNotFound):
				WriteError(w, http.StatusNotFound, "zone not found")
			default:
				WriteError(w, http.StatusInternalServerError, "internal server error")
			}
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{
			"message": "exited zone successfully",
		})
	}
}
