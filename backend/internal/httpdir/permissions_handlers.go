package httpdir

import (
	"encoding/json"
	"errors"
	"net/http"

	"tracelock/internal/access"
	"tracelock/internal/auth"
)

// POST /admin/access ; grant a user access to a zone
func GrantAccessHandler(service *access.ZoneService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		adminID, err := auth.GetUserIDFromContext(r.Context())
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req struct {
			UserID int `json:"user_id"`
			ZoneID int `json:"zone_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if req.UserID <= 0 || req.ZoneID <= 0 {
			WriteError(w, http.StatusBadRequest, "invalid user_id or zone_id")
			return
		}

		if err := service.GrantAccess(req.UserID, req.ZoneID, adminID); err != nil {
			if errors.Is(err, access.ErrZoneNotFound) {
				WriteError(w, http.StatusNotFound, "zone not found")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not grant access")
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{"message": "access granted"})
	}
}

// DELETE /admin/access ; revoke a user's access to a zone
func RevokeAccessHandler(service *access.ZoneService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req struct {
			UserID int `json:"user_id"`
			ZoneID int `json:"zone_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if req.UserID <= 0 || req.ZoneID <= 0 {
			WriteError(w, http.StatusBadRequest, "invalid user_id or zone_id")
			return
		}

		if err := service.RevokeZoneAccess(req.UserID, req.ZoneID); err != nil {
			if errors.Is(err, access.ErrAccessNotFound) {
				WriteError(w, http.StatusNotFound, "access grant not found")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not revoke access")
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{"message": "access revoked"})
	}
}

// GET /users/{id}/access ; list zones a user can enter
func ListUserAccessHandler(service *access.ZoneService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid user id")
			return
		}

		zones, err := service.ListUserAccess(userID)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "could not fetch access list")
			return
		}

		resp := make([]ZoneResponse, 0, len(zones))
		for _, z := range zones {
			resp = append(resp, ZoneResponse{
				ID:          z.ID,
				Name:        z.Name,
				Description: z.Description,
				MaxCapacity: z.MaxCapacity,
				CreatedAt:   z.CreatedAt,
			})
		}
		WriteJSON(w, http.StatusOK, resp)
	}
}

// GET /admin/zones/{id}/users; list users who have access to a zone
func ListZoneUsersHandler(service *access.ZoneService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zoneID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid zone id")
			return
		}

		users, err := service.ListZoneUsers(zoneID)
		if err != nil {
			if errors.Is(err, access.ErrZoneNotFound) {
				WriteError(w, http.StatusNotFound, "zone not found")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not fetch zone users")
			return
		}

		resp := make([]UserResponse, 0, len(users))
		for _, u := range users {
			resp = append(resp, UserResponse{
				ID:        u.ID,
				Name:      u.Name,
				Email:     u.Email,
				Role:      u.Role,
				CreatedAt: u.CreatedAt,
			})
		}
		WriteJSON(w, http.StatusOK, resp)
	}
}
