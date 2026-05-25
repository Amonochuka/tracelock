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
	device, err := s.repo.GetDevice(deviceID)
	if err != nil {
		return nil, err
	}
	// count, err := s.repo.CountActiveUsers(zoneID)
	// if err != nil {
	// 	return nil, err
	// }
	// users, err := s.repo.GetActiveUsersInZone(zoneID)
	// if err != nil {
	// 	return nil, err
	// }
	return device, nil
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
	// count, err := s.repo.CountActiveUsers(zoneID)
	// if err != nil {
	// 	return err
	// }
	// if count > 0 {
	// 	return ErrZoneHasActivity
	// }
	return s.repo.DeleteDevice(deviceID)
}

// delete a device
func (s *DeviceService) DeactivateDevice(deviceID int) error {
	// count, err := s.repo.CountActiveUsers(zoneID)
	// if err != nil {
	// 	return err
	// }
	// if count > 0 {
	// 	return ErrZoneHasActivity
	// }
	return s.repo.DeactivateDevice(deviceID)
}
