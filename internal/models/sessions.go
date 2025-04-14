package models

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/mileusna/useragent"
)

type Session struct {
	ID        int64
	UserID    int64
	Token     string
	IPAddress string
	UserAgent string
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time

	User             User
	Current          bool
	UserAgentDetails string
}

type SessionRepoInterface interface {
	Insert(s *Session) error
	Touch(s *Session) error
	Delete(s *Session) error
	DeleteByID(user User, ID int64) error
	FindByToken(token string) (Session, error)
	ListForUser(u *User) ([]Session, error)
	DeleteExpired() (int64, error)
}

type SessionRepo struct {
	db *sql.DB
}

func (r *SessionRepo) Insert(session *Session) error {
	stmt := `INSERT INTO sessions (user_id, token, ip_address, user_agent, expires_at)
	VALUES($1, $2, $3, $4, $5)
	RETURNING id;`

	args := []any{session.UserID, session.Token, session.IPAddress, session.UserAgent, session.ExpiresAt}
	err := r.db.QueryRow(stmt, args...).Scan(&session.ID)
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "idx_session_token_unique"` {
			return ErrDuplicateToken
		}

		return fmt.Errorf("[SessionRepo.Insert] %w", err)
	}

	return nil
}

func (r *SessionRepo) Touch(session *Session) error {
	stmt := `UPDATE sessions SET updated_at = (NOW() AT TIME ZONE 'UTC') WHERE id = $1;`
	_, err := r.db.Exec(stmt, session.ID)
	if err != nil {
		return fmt.Errorf("[SessionRepo.Touch] %s", err)
	}

	return nil
}

func (r *SessionRepo) Delete(session *Session) error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE token = $1;", session.Token)
	if err != nil {
		return fmt.Errorf("[SessionRepo.Delete] %w", err)
	}

	return nil
}

func (r *SessionRepo) DeleteByID(user User, id int64) error {
	result, err := r.db.Exec("DELETE FROM sessions WHERE user_id = $1 AND id = $2;", user.ID, id)
	if err != nil {
		return fmt.Errorf("[SessionRepo.DeleteByID] %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("[SessionRepo.DeleteByID] %w", err)
	}

	if affectedRows == 0 {
		return ErrNoRecord
	}

	return nil
}

func (r *SessionRepo) FindByToken(token string) (Session, error) {
	stmt := `SELECT sessions.*, users.*
	FROM sessions
	INNER JOIN users ON sessions.user_id = users.id
	WHERE sessions.token = $1
		AND sessions.expires_at >= (NOW() AT TIME ZONE 'UTC');`

	var s Session
	err := r.db.QueryRow(stmt, token).Scan(
		&s.ID,
		&s.UserID,
		&s.Token,
		&s.IPAddress,
		&s.UserAgent,
		&s.ExpiresAt,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.User.ID,
		&s.User.Email,
		&s.User.HashedPassword,
		&s.User.Timezone,
		&s.User.CreatedAt,
		&s.User.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s, ErrNoRecord
		}

		return s, fmt.Errorf("[SessionRepo.FindByToken] %w", err)
	}

	return s, nil
}

func (r *SessionRepo) ListForUser(u *User) ([]Session, error) {
	stmt := `
		SELECT id, token, ip_address, user_agent, expires_at, created_at, updated_at
		FROM sessions
		WHERE user_id = $1
		ORDER BY updated_at DESC;`

	var sessions []Session
	rows, err := r.db.Query(stmt, u.ID)
	if err != nil {
		return sessions, fmt.Errorf("[SessionRepo.ListForUser] %w", err)
	}

	for rows.Next() {
		var s Session
		err = rows.Scan(&s.ID, &s.Token, &s.IPAddress, &s.UserAgent, &s.ExpiresAt, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return sessions, fmt.Errorf("[SessionRepo.ListForUser] %w", err)
		}

		ua := useragent.Parse(s.UserAgent)
		uaDetails := []string{ua.Name, ua.Device, ua.OS}
		uaDetails = slices.DeleteFunc(uaDetails, func(item string) bool { return item == "" })
		s.UserAgentDetails = strings.Join(uaDetails, " / ")

		sessions = append(sessions, s)
	}

	if err = rows.Err(); err != nil {
		return sessions, fmt.Errorf("[SessionRepo.ListForUser] %w", err)
	}

	return sessions, nil
}

func (r *SessionRepo) DeleteExpired() (int64, error) {
	result, err := r.db.Exec("DELETE FROM sessions WHERE expires_at <= (NOW() AT TIME ZONE 'UTC');")
	if err != nil {
		return 0, nil
	}

	deletedRecords, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return deletedRecords, nil
}
