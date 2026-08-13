package access

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type TicketInfo struct {
	UserID    int
	Role      string
	ExpiresAt time.Time
}

type TicketStore struct {
	tickets map[string]TicketInfo
	mu      sync.Mutex
}

func NewTicketStore() *TicketStore {
	store := &TicketStore{
		tickets: make(map[string]TicketInfo),
	}
	// start cleanup routine
	go store.cleanupRoutine()
	return store
}

func (s *TicketStore) GenerateTicket(userID int, role string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	ticket := hex.EncodeToString(bytes)

	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Valid for 15 seconds
	s.tickets[ticket] = TicketInfo{
		UserID:    userID,
		Role:      role,
		ExpiresAt: time.Now().Add(15 * time.Second),
	}
	return ticket, nil
}

func (s *TicketStore) ConsumeTicket(ticket string) (TicketInfo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, exists := s.tickets[ticket]
	if !exists {
		return TicketInfo{}, false
	}
	
	// Delete immediately so it can only be used once
	delete(s.tickets, ticket)

	if time.Now().After(info.ExpiresAt) {
		return TicketInfo{}, false
	}

	return info, true
}

func (s *TicketStore) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for ticket, info := range s.tickets {
			if now.After(info.ExpiresAt) {
				delete(s.tickets, ticket)
			}
		}
		s.mu.Unlock()
	}
}
