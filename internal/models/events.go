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
