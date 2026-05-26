package access

import "tracelock/internal/models"

type CredentialService struct {
	repo *CredentialRepo
}

func NewCredentialService(repo *CredentialRepo) *CredentialService {
	return &CredentialService{repo: repo}
}

func (s *CredentialService) EnrollCredential(userID int, entryMethod, credentialHash string) (*models.BiometricCredential, error) {
	return s.repo.EnrollCredential(userID, entryMethod, credentialHash)
}

func (s *CredentialService) GetCredential(userID int, entryMethod string) (*models.BiometricCredential, error) {
	return s.repo.GetCredential(userID, entryMethod)
}

func (s *CredentialService) RevokeCredential(userID int, entryMethod string) error {
	return s.repo.RevokeCredential(userID, entryMethod)
}

func (s *CredentialService) ListUserCredentials(userID int) ([]*models.BiometricCredential, error) {
	return s.repo.ListUserCredentials(userID)
}
