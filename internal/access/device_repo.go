package access

import "tracelock/internal/models"

type DeviceRepo struct{
	db *sql.DB
}

func NewDeviceRepo(db *sql.DB) *DeviceRepo{
	return &DeviceRepo{db: db}
}

func (d *DeviceRepo) CreateDevice(zoneID int, name, deviceType, serial string) (*models.Device, error) {
	device := &models.Device{}
	err := d.db.QueryRow(`INSERT INTO devices(zone_id, name, type, serial)
		VALUES($1,$2,$3,$4) RETURNING id, zone_id, name, type, serial, active, created_at`,
		zoneID, name, deviceType, serial).
		Scan(&device.ID, &device.ZoneID, &device.Name, &device.Type, &device.Serial, &device.Active, &device.CreatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrDeviceSerialExists
		}
		return nil, fmt.Errorf("create device: %w", err)
	}
	return device, nil
}

func (d *DeviceRepo) GetDevice(deviceID int) (*models.Device, error) {
	dvc := &models.Device{}
	err := d.db.QueryRow("SELECT id, zone_id, name, type, serial, active, created_at FROM devices WHERE id = $1", deviceID).
		Scan(&dvc.ID, &dvc.ZoneID, &dvc.Name, &dvc.Type, &dvc.Serial, &dvc.Active, &dvc.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDeviceNotFound
		}
		return nil, fmt.Errorf("get device: %w", err)
	}
	return dvc, nil
}

func (d *DeviceRepo) ListZoneDevices(zoneID int) ([]*models.Device, error) {
	rows, err := d.db.Query(`SELECT id, zone_id, name, type, serial, active, created_at FROM devices WHERE zone_id = $1`, zoneID)
	if err != nil {
		return nil, fmt.Errorf("list zone devices: %w", err)
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		dvc := &models.Device{}
		if err := rows.Scan(&dvc.ID, &dvc.ZoneID, &dvc.Name, &dvc.Type, &dvc.Serial, &dvc.Active, &dvc.CreatedAt); err != nil {
			return nil, fmt.Errorf("device listing: %w", err)
		}
		devices = append(devices, dvc)
	}
	return devices, nil
}