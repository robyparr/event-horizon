package mocks

import (
	"slices"
	"time"

	"github.com/robyparr/event-horizon/internal/models"
)

type UserRepo struct {
	users map[int64]models.User
}

func NewUserRepo() *UserRepo {
	return &UserRepo{users: make(map[int64]models.User)}
}

func (r *UserRepo) Insert(u *models.User) error {
	for _, user := range r.users {
		if user.Email == u.Email {
			return models.ErrDuplicateEmail
		}
	}

	u.ID = int64(len(r.users) + 1)
	u.HashedPassword = []byte(u.Password)
	u.CreatedAt = time.Now().UTC()
	u.UpdatedAt = time.Now().UTC()

	r.users[u.ID] = *u
	return nil
}

func (r *UserRepo) Authenticate(email string, password string) (int64, error) {
	for id, user := range r.users {
		if user.Email == email && slices.Equal(user.HashedPassword, []byte(password)) {
			return id, nil
		}
	}

	return 0, models.ErrInvalidCredentials
}
