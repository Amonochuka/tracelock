package httpdir

import (
	"net/http"
	"tracelock/internal/access"
)

func EnrollCredentialHandler(service *access.CredentialService) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var req struct {
			Name   string `json:"name"`
			Type   string `json:"type"`
			Serial string `json:"serial"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		zoneID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid zone id")
			return
		}

		device, err := service.CreateDevice(zoneID, req.Name, req.Type, req.Serial)
		if err != nil {
			if errors.Is(err, access.ErrDeviceSerialExists) {
				WriteError(w, http.StatusConflict, "device already exists")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not create device")
			return
		}
		WriteJSON(w, http.StatusCreated, DeviceResponse{
			ID:        device.ID,
			ZoneID:    device.ZoneID,
			Name:      device.Name,
			Type:      device.Type,
			Serial:    device.Serial,
			Active:    device.Active,
			CreatedAt: device.CreatedAt,
		})
	}
}

func GetCredentialHandler(service *access.CredentialService)http.HandlerFunc{}
func RevokeCredentialHandler(service *access.CredentialService)http.HandlerFunc{}
func ListUserCredentialsHandler(service *access.CredentialService) http.HandlerFunc{}