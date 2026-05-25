package httpdir

import (
	"encoding/json"
	"errors"
	"net/http"
	"tracelock/internal/access"
)

func CreateDeviceHandler(service *access.DeviceService) http.HandlerFunc {
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
			Name:      device.Name,
			Type:      device.Type,
			Active:    device.Active,
			CreatedAt: device.CreatedAt,
		})
	}
}

func GetDeviceHandler(service *access.DeviceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deviceID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid device id")
			return
		}

		device, err := service.GetDevice(deviceID)
		if err != nil {
			if errors.Is(err, access.ErrDeviceNotFound) {
				WriteError(w, http.StatusNotFound, "device not found")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not fetch device")
			return
		}
		WriteJSON(w, http.StatusOK, device)
	}
}

func ListDevicesHandler(service *access.DeviceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		zoneID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid zone id")
			return
		}

		devices, err := service.ListDevices(zoneID)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "could not fetch devices")
			return
		}
		WriteJSON(w, http.StatusOK, devices)
	}
}

func UpdateDeviceHandler(service *access.DeviceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		deviceID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid device id")
			return
		}

		var req struct {
			Name   string `json:"name"`
			Type   string `json:"type"`
			Serial string `json:"serial"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		device, err := service.UpdateDevice(deviceID, req.Name, req.Type, req.Serial)
		if err != nil {
			if errors.Is(err, access.ErrDeviceNotFound) {
				WriteError(w, http.StatusNotFound, "device not found")
				return
			}
			WriteError(w, http.StatusInternalServerError, "could not update device")
			return
		}

		WriteJSON(w, http.StatusOK, DeviceResponse{
			ID:        device.ID,
			Name:      device.Name,
			Type:      device.Type,
			Active:    device.Active,
			CreatedAt: device.CreatedAt,
		})
	}
}

func DeactivateDeviceHandler(service *access.DeviceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deviceID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid device id")
			return
		}

		if err := service.DeactivateDevice(deviceID); err != nil {
			switch {
			case errors.Is(err, access.ErrDeviceNotFound):
				WriteError(w, http.StatusNotFound, "device not found")
			default:
				WriteError(w, http.StatusInternalServerError, "could not deactivate device")
			}
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{"message": "device deactivated"})
	}
}

func DeleteDeviceHandler(service *access.DeviceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deviceID, err := parseIDParam(r, "id")
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid device id")
			return
		}

		if err := service.DeleteDevice(deviceID); err != nil {
			switch {
			case errors.Is(err, access.ErrDeviceNotFound):
				WriteError(w, http.StatusNotFound, "device not found")
			default:
				WriteError(w, http.StatusInternalServerError, "could not delete device")
			}
			return
		}

		WriteJSON(w, http.StatusOK, map[string]string{"message": "device deleted"})
	}
}
