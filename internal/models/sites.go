package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/robyparr/event-horizon/internal/utils"
)

type Site struct {
	ID        int64
	UserID    int64
	Name      string
	Token     string
	CreatedAt time.Time
	UpdatedAt time.Time

	EventCount int
}

type SiteRepoInterface interface {
	ListForUser(u *User) ([]Site, error)
	Insert(s *Site) error
	FindForUser(u *User, id int64) (Site, error)
	Delete(s *Site) error
	FindByToken(token string) (Site, error)
}

type SiteRepo struct {
	db *sql.DB
}

func (r *SiteRepo) ListForUser(u *User) ([]Site, error) {
	var sites []Site
	stmt := `
		SELECT *, (SELECT COUNT(*) FROM events WHERE site_id = sites.id) FROM sites
		WHERE sites.user_id = $1
		;
	`
	rows, err := r.db.Query(stmt, u.ID)
	if err != nil {
		return sites, fmt.Errorf("[SiteRepo.ListForUser] %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var s Site

		err = rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Token, &s.CreatedAt, &s.UpdatedAt, &s.EventCount)
		if err != nil {
			return sites, fmt.Errorf("[SiteRepo.ListForUser] %w", err)
		}

		sites = append(sites, s)
	}

	if err = rows.Err(); err != nil {
		return sites, fmt.Errorf("[SiteRepo.ListForUser] %w", err)
	}

	return sites, nil
}

func (r *SiteRepo) Insert(s *Site) error {
	stmt := `INSERT INTO sites (user_id, name, token)
	VALUES($1, $2, $3) RETURNING id, created_at, updated_at;`

	token := utils.Token()
	err := r.db.QueryRow(stmt, s.UserID, s.Name, token).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("[SiteRepo.Insert] %w", err)
	}

	s.Token = token
	return nil
}

func (r *SiteRepo) FindForUser(u *User, id int64) (Site, error) {
	var s Site
	stmt := `
SELECT *,
	(SELECT COUNT(*) FROM events WHERE site_id = sites.id)
FROM sites
WHERE user_id = $1
	AND id = $2;
`

	err := r.db.QueryRow(stmt, u.ID, id).Scan(
		&s.ID,
		&s.UserID,
		&s.Name,
		&s.Token,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.EventCount,
	)

	if err != nil {
		return s, fmt.Errorf("[SiteRepo.FindForUser] %w", err)
	}

	return s, nil
}

func (r *SiteRepo) Delete(s *Site) error {
	_, err := r.db.Exec("DELETE FROM sites WHERE id = $1;", s.ID)
	if err != nil {
		return fmt.Errorf("[SiteRepo.Delete] %w", err)
	}

	return nil
}

func (r *SiteRepo) FindByToken(token string) (Site, error) {
	var s Site
	err := r.db.QueryRow("SELECT * FROM sites WHERE token = $1 LIMIT 1;", token).Scan(
		&s.ID,
		&s.UserID,
		&s.Name,
		&s.Token,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		return s, fmt.Errorf("[SiteRepo.FindByToken] %w", err)
	}

	return s, nil
}
