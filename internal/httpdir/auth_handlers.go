package httpdir

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"tracelock/internal/access"
	"tracelock/internal/auth"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *RegisterRequest) Validate() error {
	if len(strings.TrimSpace(r.Name)) < 2 {
		return errors.New("name must be at least two characters")
	}
	if !strings.Contains(r.Email, "@") {
		return errors.New("invalid email")
	}
	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

func (l *LoginRequest) Validate() error {
	if !strings.Contains(l.Email, "@") {
		return errors.New("invalid email")
	}
	if len(l.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

func RegisterHandler(s *auth.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req RegisterRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if err := req.Validate(); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := s.Register(req.Name, req.Email, req.Password); err != nil {
			if errors.Is(err, auth.ErrEmailExists) {
				WriteError(w, http.StatusConflict, "email already exists")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not register user")
			return
		}

		WriteJSON(w, http.StatusCreated, map[string]string{
			"message": "user registered successfully",
		})
	}
}

func LoginHandler(s *auth.UserService, j *auth.JWTService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req LoginRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if err := req.Validate(); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		user, err := s.Authenticate(req.Email, req.Password)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) {
				WriteError(w, http.StatusUnauthorized, "invalid credentials")
				return
			}
			WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		token, err := j.GenerateToken(user)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "could not generate token")
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{
			"token": token,
		})
	}
}

func MeHandler(s *auth.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := auth.GetUserClaims(r)
		if claims == nil {
			WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		user, err := s.VerifyUser(claims.UserID)
		if err != nil {
			if errors.Is(err, auth.ErrUserNotFound) {
				WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not fetch user")
			return
		}

		WriteJSON(w, http.StatusOK, UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		})
	}
}

// bootstrap handler
func BootStrapHandler(s *auth.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req RegisterRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}

		if err := req.Validate(); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.BootStrapAdmin(req.Name, req.Email, req.Password); err != nil {
			if errors.Is(err, auth.ErrAdminExists) {
				WriteError(w, http.StatusForbidden, "admin already exists")
				return
			}
			if errors.Is(err, auth.ErrEmailExists) {
				WriteError(w, http.StatusConflict, "email already exists")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not create admin")
			return
		}
		WriteJSON(w, http.StatusCreated, map[string]string{
			"message": "admin account created",
		})
	}
}

func ListUsersHandler(s *auth.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := s.ListUsers()
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "could not fetch users")
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

func UpdateRoleHandler(s *auth.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		targetID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		var req struct {
			Role string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if err := s.UpdateRole(targetID, req.Role); err != nil {
			switch {
			case errors.Is(err, auth.ErrUserNotFound):
				WriteError(w, http.StatusNotFound, "user not found")
			case errors.Is(err, auth.ErrInvalidRole):
				WriteError(w, http.StatusNotFound, "role must be 'admin' or 'user'")
			default:
				WriteError(w, http.StatusInternalServerError, "could not update role")
			}
			return
		}
		WriteJSON(w, http.StatusOK, map[string]string{
			"message": "role updated successfully",
		})
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

func ListUserEventsHandler(service *access.ZoneService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid user id")
			return
		}

		limit, offset := parsePagination(r)

		events, total, err := service.ListUserEvents(userID, limit, offset)
		if err != nil {
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
