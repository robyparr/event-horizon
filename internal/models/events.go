package models

import (
	"cmp"
	"database/sql"
	"encoding/json"
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
	MetricCounts(site *Site) (EventMetrics, error)
}

type EventMetrics struct {
	DeviceType map[string]int
	OS         map[string]int
	Browser    map[string]int
}

func (em *EventMetrics) ToJSON() (map[string]string, error) {
	deviceTypeJSON, deviceTypeErr := json.Marshal(em.DeviceType)
	osJSON, osErr := json.Marshal(em.OS)
	browserJSON, browserErr := json.Marshal(em.Browser)

	firstError := cmp.Or(deviceTypeErr, osErr, browserErr)
	if firstError != nil {
		return nil, firstError
	}

	return map[string]string{
		"deviceType": string(deviceTypeJSON),
		"os":         string(osJSON),
		"browser":    string(browserJSON),
	}, nil
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

func (r *EventRepo) MetricCounts(site *Site) (EventMetrics, error) {
	days := 7
	endOn := time.Now().UTC().Truncate(24 * time.Hour)
	startOn := endOn.AddDate(0, 0, -(days - 1))

	stmt := `
WITH filtered_events AS (
	SELECT * FROM events
	WHERE site_id = $1
		AND created_at::DATE BETWEEN $2 AND $3
)
SELECT device_type, count(*), 'device_type' AS metric
FROM filtered_events
GROUP BY device_type
UNION
SELECT os, count(*), 'os' AS metric
FROM filtered_events
GROUP BY os
UNION
SELECT browser, count(*), 'browser' AS metric
FROM filtered_events
GROUP BY browser
`

	out := EventMetrics{
		DeviceType: make(map[string]int),
		OS:         make(map[string]int),
		Browser:    make(map[string]int),
	}

	rows, err := r.db.Query(stmt, site.ID, startOn, endOn)
	if err != nil {
		return out, fmt.Errorf("[EventRepo.MetricCounts] %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var value string
		var count int
		var metric string

		err := rows.Scan(&value, &count, &metric)
		if err != nil {
			return out, fmt.Errorf("[EventRepo.MetricCounts] %w", err)
		}

		switch metric {
		case "device_type":
			out.DeviceType[value] += 1
		case "os":
			out.OS[value] += 1
		case "browser":
			out.Browser[value] += 1
		default:
			panic("Unknown metric: " + metric)
		}
	}

	if err := rows.Err(); err != nil {
		return out, fmt.Errorf("[EventRepo.MetricCounts] %w", err)
	}

	return out, nil
}
