package models

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserRepoInterface interface {
	Insert(u *User) error
	Authenticate(email string, password string) (int64, error)
}

type User struct {
	ID             int64
	Name           string
	Email          string
	HashedPassword []byte
	Timezone       string
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Password string
}

type UserRepo struct {
	db *sql.DB
}

func (r *UserRepo) Insert(user *User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		return err
	}

	stmt := "INSERT INTO users (email, hashed_password, timezone) VALUES($1, $2, $3) RETURNING ID;"
	err = r.db.QueryRow(stmt, user.Email, string(hashedPassword), user.Timezone).Scan(&user.ID)

	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "idx_users_email_unique"` {
			return ErrDuplicateEmail
		}

		return err
	}

	user.HashedPassword = hashedPassword
	return nil
}

func (r *UserRepo) Authenticate(email string, password string) (int64, error) {
	var id int64
	var hashedPassword []byte

	stmt := "SELECT id, hashed_password from users WHERE email = $1"
	err := r.db.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil
}
