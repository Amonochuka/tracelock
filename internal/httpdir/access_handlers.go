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

func CreateZoneHandler(service *access.ZoneService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			MaxCapacity int    `json:"max_capacity"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Name == "" {
			WriteError(w, http.StatusBadRequest, "name is required")
			return
		}

		zone, err := service.CreateZone(req.Name, req.Description, req.MaxCapacity)
		if err != nil {
			if errors.Is(err, access.ErrZoneNameExists) {
				WriteError(w, http.StatusConflict, "zone name already exists")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not create zone")
			return
		}
		WriteJSON(w, http.StatusCreated, ZoneResponse{
			ID:          zone.ID,
			Name:        zone.Name,
			Description: zone.Description,
			MaxCapacity: zone.MaxCapacity,
			CreatedAt:   zone.CreatedAt,
		})
	}
}

func GetZoneHandler(service *access.ZoneService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zoneID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid zone id")
			return
		}

		zone, err := service.GetZone(zoneID)
		if err != nil {
			if errors.Is(err, access.ErrZoneNotFound) {
				WriteError(w, http.StatusNotFound, "zone not found")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not fetch zone")
			return
		}
		WriteJSON(w, http.StatusOK, zone)
	}
}

func ListZonesHandler(service *access.ZoneService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zones, err := service.ListZones()
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "could not fetch zones")
			return
		}
		WriteJSON(w, http.StatusOK, zones)
	}
}

func ListZoneEventsHandler(service *access.ZoneService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zoneID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid zone id")
			return
		}

		limit, offset := parsePagination(r)

		events, total, err := service.ListZoneEvents(zoneID, limit, offset)
		if err != nil {
			if errors.Is(err, access.ErrZoneNotFound) {
				WriteError(w, http.StatusNotFound, "zone not found")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not fetch events")
			return
		}

		WriteJSON(w, http.StatusOK, map[string]any{
			"events": events,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		})
	}
}

func UpdateZoneHandler(service *access.ZoneService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		zoneID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid zone id")
			return
		}

		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			MaxCapacity int    `json:"max_capacity"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.Name == "" {
			WriteError(w, http.StatusBadRequest, "name is required")
			return
		}

		zone, err := service.UpdateZone(zoneID, req.Name, req.Description, req.MaxCapacity)
		if err != nil {
			if errors.Is(err, access.ErrZoneNotFound) {
				WriteError(w, http.StatusNotFound, "zone not found")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not update zone")
			return
		}

		WriteJSON(w, http.StatusOK, ZoneResponse{
			ID:          zone.ID,
			Name:        zone.Name,
			Description: zone.Description,
			MaxCapacity: zone.MaxCapacity,
			CreatedAt:   zone.CreatedAt,
		})
	}
}

func DeleteZoneHandler(service *access.ZoneService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zoneID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid zone id")
			return
		}

		if err := service.DeleteZone(zoneID); err != nil {
			switch {
			case errors.Is(err, access.ErrZoneNotFound):
				WriteError(w, http.StatusNotFound, "zone not found")
			case errors.Is(err, access.ErrZoneHasActivity):
				WriteError(w, http.StatusConflict, "zone has active sessions and cannot be deleted")
			default:
				WriteError(w, http.StatusInternalServerError, "could not delete zone")
			}
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{"message": "zone deleted"})
	}
}

func VerifyChainHandler(service *access.ZoneService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zoneID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid zone id")
			return
		}

		valid, count, err := service.VerifyChain(zoneID)
		if err != nil {
			if errors.Is(err, access.ErrZoneNotFound) {
				WriteError(w, http.StatusNotFound, "zone not found")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not verify chain")
			return
		}

		msg := "chain is intact"
		if !valid {
			msg = "chain integrity violation detected"
		}

		WriteJSON(w, http.StatusOK, map[string]any{
			"zone_id":        zoneID,
			"valid":          valid,
			"events_checked": count,
			"message":        msg,
		})
	}
}
