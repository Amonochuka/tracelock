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

		timestamp := time.Now()

		err = service.HandleZoneEvent(userID, req.ZoneID, "enter", timestamp)
		if err != nil {
			switch {
			case errors.Is(err, access.ErrZoneFull):
				WriteError(w, http.StatusConflict, err.Error())

			case errors.Is(err, access.ErrUserAlreadyInZone):
				WriteError(w, http.StatusConflict, err.Error())

			case errors.Is(err, access.ErrNoActiveSession):
				WriteError(w, http.StatusBadRequest, err.Error())

			case errors.Is(err, access.ErrZoneNotFound):
				WriteError(w, http.StatusNotFound, err.Error())

			default:
				WriteError(w, http.StatusInternalServerError, "internal server error")
			}

			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{
			"message": "entered successfully",
		})
	}
}
