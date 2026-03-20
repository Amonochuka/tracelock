package http

import (
	"encoding/json"
	"net/http"
	"tracelock/internal/access"
	"tracelock/internal/auth"
)

func EnterZoneHandler(service *access.ZoneService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.GetUserIDFromContext(r.Context())
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var req struct {
			ZoneID int `json:"zone_id"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		err = service.EnterZone(userID, req.ZoneID)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		w.Write([]byte("entered successfully"))
	}

}
