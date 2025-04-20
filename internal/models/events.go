package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Event struct {
	ID         int64
	SiteID     int64
	Action     string
	Count      int
	DeviceType string
	OS         string
	Browser    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type EventRepoInterface interface {
	Insert(event *Event) error
	CountsByDate(site *Site) (map[string]int, error)
}

type EventRepo struct {
	db *sql.DB
}

func (r *EventRepo) Insert(event *Event) error {
	stmt := `INSERT INTO events (site_id, action, count, device_type, os, browser)
	VALUES($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at;`

	err := r.db.
		QueryRow(stmt, event.SiteID, event.Action, event.Count, event.DeviceType, event.OS, event.Browser).
		Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)

	if err != nil {
		return fmt.Errorf("[EventRepo.Insert] %w", err)
	}

	return nil
}

func (r *EventRepo) CountsByDate(site *Site) (map[string]int, error) {
	days := 7
	endOn := time.Now().UTC().Truncate(24 * time.Hour)
	startOn := endOn.AddDate(0, 0, -(days - 1))

	out := make(map[string]int, days)
	for d := range days {
		t := startOn.AddDate(0, 0, d)
		out[t.Format("2006-01-02")] = 0
	}

	stmt := `
		SELECT DATE_TRUNC('day', created_at)::DATE, COUNT(*)
		FROM events
		WHERE site_id = $1
			AND created_at::DATE BETWEEN $2::DATE AND $3::DATE
		GROUP BY DATE_TRUNC('day', created_at)
		ORDER BY DATE_TRUNC('day', created_at);
	`

	rows, err := r.db.Query(stmt, site.ID, startOn, endOn)
	if err != nil {
		return out, fmt.Errorf("[EventRepo.CountsByDate] %w", err)
	}

	defer rows.Close()
	for rows.Next() {
		var date time.Time
		var count int

		err := rows.Scan(&date, &count)
		if err != nil {
			return out, fmt.Errorf("[EventRepo.CountsByDate] %w", err)
		}

		out[date.Format("2006-01-02")] = count
	}

	if err = rows.Err(); err != nil {
		return out, fmt.Errorf("[EventRepo.CountsByDate] %w", err)
	}

	return out, nil
}
