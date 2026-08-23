package auth

import (
	"errors"
	"testing"
	"time"

	"tracelock/internal/models"
)

type fakeUserRepo struct {
	users        map[int]*models.User
	otherAdmins  int
	updateCalled bool
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: map[int]*models.User{}}
}

func (f *fakeUserRepo) add(u *models.User) { f.users[u.ID] = u }

func (f *fakeUserRepo) Register(name, email, password string) error { return nil }
func (f *fakeUserRepo) Authenticate(email, password string) (*models.User, error) {
	return nil, ErrInvalidCredentials
}
func (f *fakeUserRepo) VerifyUser(id int) (*models.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}
func (f *fakeUserRepo) AdminExists() (bool, error)                       { return false, nil }
func (f *fakeUserRepo) RegisterAdmin(name, email, password string) error { return nil }
func (f *fakeUserRepo) ResetAdminPassword(email, password string) error  { return nil }
func (f *fakeUserRepo) UpdateRole(userID int, role string) error {
	f.updateCalled = true
	f.users[userID].Role = role
	return nil
}
func (f *fakeUserRepo) ListUsers() ([]*models.User, error) { return nil, nil }
func (f *fakeUserRepo) SaveRefreshToken(userID int, token string, expiresAt time.Time) error {
	return nil
}
func (f *fakeUserRepo) ValidateAndGetUserIDFromRefreshToken(tokenHash string) (int, error) {
	return 0, nil
}
func (f *fakeUserRepo) RevokeRefreshToken(token string) error { return nil }
func (f *fakeUserRepo) DeleteExpiredTokens() error            { return nil }
func (f *fakeUserRepo) IncrementFailedAttempts(email string) (int, error) {
	return 0, nil
}
func (f *fakeUserRepo) LockAccount(email string) error         { return nil }
func (f *fakeUserRepo) ResetFailedAttempts(email string) error { return nil }
func (f *fakeUserRepo) IsAccountLocked(email string) (bool, error) {
	return false, nil
}
func (f *fakeUserRepo) UnlockAccount(userID int) error { return nil }
func (f *fakeUserRepo) DeleteUser(userID int) error    { return nil }
func (f *fakeUserRepo) CountOtherActiveAdmins(excludeUserID int) (int, error) {
	return f.otherAdmins, nil
}
func (f *fakeUserRepo) EnsureDemoAdmin(email, name, password string) (string, error) {
	return "ok", nil
}

func TestUpdateRoleBlocksDemotingLastAdmin(t *testing.T) {
	repo := newFakeUserRepo()
	repo.add(&models.User{ID: 1, Role: "admin"})
	repo.otherAdmins = 0
	svc := NewUserService(repo, nil)

	err := svc.UpdateRole(1, "user")
	if !errors.Is(err, ErrLastAdmin) {
		t.Fatalf("expected ErrLastAdmin, got %v", err)
	}
	if repo.updateCalled {
		t.Fatal("repo.UpdateRole must not be called when blocking")
	}
}

func TestUpdateRoleAllowsDemoteWhenOtherAdminExists(t *testing.T) {
	repo := newFakeUserRepo()
	repo.add(&models.User{ID: 1, Role: "admin"})
	repo.add(&models.User{ID: 2, Role: "admin"})
	repo.otherAdmins = 1
	svc := NewUserService(repo, nil)

	if err := svc.UpdateRole(1, "user"); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !repo.updateCalled || repo.users[1].Role != "user" {
		t.Fatal("demotion should have been applied")
	}
}

func TestUpdateRoleAllowsPromoteWithoutAdminCount(t *testing.T) {
	repo := newFakeUserRepo()
	repo.add(&models.User{ID: 3, Role: "user"})
	svc := NewUserService(repo, nil)

	if err := svc.UpdateRole(3, "admin"); err != nil {
		t.Fatalf("expected promotion to succeed, got %v", err)
	}
}

func TestUpdateRoleRejectsInvalidRole(t *testing.T) {
	svc := NewUserService(newFakeUserRepo(), nil)
	if err := svc.UpdateRole(1, "superadmin"); !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("expected ErrInvalidRole, got %v", err)
	}
}
