package access

import "tracelock/internal/models"

type DeviceService struct {
	repo *DeviceRepo
}

func NewDeviceService(repo *DeviceRepo) *DeviceService {
	return &DeviceService{repo: repo}
}

// --device management--
// list all existing devices
func (s *DeviceService) ListDevices(zoneID int) ([]*models.Device, error) {
	return s.repo.ListZoneDevices(zoneID)
}

// get a particular device
func (s *DeviceService) GetDevice(deviceID int) (*models.Device, error) {
    return s.repo.GetDevice(deviceID)
}

// create a new device
func (s *DeviceService) CreateDevice(zoneID int, name, deviceType, serial string) (*models.Device, error) {
	return s.repo.CreateDevice(zoneID, name, deviceType, serial)
}

// update a device's details
func (s *DeviceService) UpdateDevice(deviceID int, name, deviceType, serial string) (*models.Device, error) {
	return s.repo.UpdateDevice(deviceID, name, deviceType, serial)
}

func (s *DeviceService) DeleteDevice(deviceID int) error {
	return s.repo.DeleteDevice(deviceID)
}

// delete a device
func (s *DeviceService) DeactivateDevice(deviceID int) error {
	return s.repo.DeactivateDevice(deviceID)
}
