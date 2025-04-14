package mocks

import (
	"fmt"
	"time"

	"github.com/robyparr/event-horizon/internal/models"
)

type SessionRepo struct {
	Sessions map[int64]models.Session
}

func NewSessionRepo() *SessionRepo {
	return &SessionRepo{Sessions: make(map[int64]models.Session)}
}

func (r *SessionRepo) Insert(s *models.Session) error {
	if s.UserID == 0 {
		return fmt.Errorf("Invalid UserID of 0")
	}

	s.ID = int64(len(r.Sessions) + 1)
	s.CreatedAt = time.Now().UTC()
	s.UpdatedAt = s.CreatedAt

	r.Sessions[s.ID] = *s
	return nil
}

func (r *SessionRepo) Touch(s *models.Session) error {
	_, ok := r.Sessions[s.ID]
	if !ok {
		return fmt.Errorf("No session with ID %d", s.ID)
	}

	s.UpdatedAt = time.Now().UTC()
	r.Sessions[s.ID] = *s
	return nil
}

func (r *SessionRepo) Delete(s *models.Session) error {
	_, ok := r.Sessions[s.ID]
	if !ok {
		return fmt.Errorf("No session with ID %d", s.ID)
	}

	delete(r.Sessions, s.ID)
	return nil
}

func (r *SessionRepo) DeleteByID(user models.User, id int64) error {
	s, ok := r.Sessions[id]
	if !ok || s.UserID != user.ID {
		return models.ErrNoRecord
	}

	delete(r.Sessions, id)
	return nil
}

func (r *SessionRepo) FindByToken(token string) (models.Session, error) {
	for _, s := range r.Sessions {
		if s.Token == token {
			return s, nil
		}
	}

	return models.Session{}, models.ErrNoRecord
}

func (r *SessionRepo) ListForUser(user *models.User) ([]models.Session, error) {
	var out []models.Session
	for _, s := range r.Sessions {
		if s.UserID == user.ID {
			out = append(out, s)
		}
	}

	if len(out) == 0 {
		return out, models.ErrNoRecord
	}

	return out, nil
}

func (r *SessionRepo) DeleteExpired() (int64, error) {
	count := 0
	for _, s := range r.Sessions {
		if s.ExpiresAt.UTC().Before(time.Now().UTC()) {
			delete(r.Sessions, s.ID)
			count += 1
		}
	}

	return int64(count), nil
}
