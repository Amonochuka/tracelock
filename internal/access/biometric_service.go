package access

type BiometricService struct {
	credentials *CredentialRepo
	devices     *DeviceRepo
	zones       *ZoneRepo
}

func NewBiometricService(credentials *CredentialRepo, devices *DeviceRepo, zones *ZoneRepo) *BiometricService {
	return &BiometricService{
		credentials: credentials,
		devices:     devices,
		zones:       zones,
	}
}