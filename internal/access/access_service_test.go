package access_test

import (
	"errors"
	"testing"
	"tracelock/internal/access"
)

type mockRepo struct {
	access bool
	err    error
}

func (r *mockRepo) HasZoneAccess(userId, zoneId int) (bool, error) {
	return r.access, r.err
}

// added wrapper, so can inject mockRepo
type testZoneService struct {
	repo interface {
		HasZoneAccess(userID, zoneID int) (bool, error)
	}
}

func (s *testZoneService) CanEnterRoom(userID, zoneID int) error {
	ok, err := s.repo.HasZoneAccess(userID, zoneID)
	if err != nil {
		return err
	}
	if !ok {
		return access.ErrAccessDenied
	}
	return nil
}

func TestCanEnterRoom_Allowed(t *testing.T) {
	svc := &testZoneService{repo: &mockRepo{access: true}}
	if err := svc.CanEnterRoom(1, 1); err != nil {
		t.Fatalf("expected nil but got %v", err)
	}
}

func TestCanEnterRoom_Denied(t *testing.T) {
	svc := &testZoneService{repo: &mockRepo{access: false}}
	err := svc.CanEnterRoom(1, 1)
	if !errors.Is(err, access.ErrAccessDenied) {
		t.Fatalf("expected ErrAccessDenied but got %v", err)
	}
}

func TestCanEnterRoom_RepoError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := &testZoneService{repo: &mockRepo{err: repoErr}}
	err := svc.CanEnterRoom(1, 1)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error but got %v", err)
	}
}
