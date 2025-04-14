package views

import (
	"testing"
	"time"

	"github.com/robyparr/event-horizon/internal/assert"
)

func TestHumanDate(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		tz   string
		want string
	}{
		{
			name: "UTC",
			t:    time.Date(2024, 3, 17, 10, 15, 0, 0, time.UTC),
			tz:   "UTC",
			want: "Mar 17, 2024 UTC",
		},
		{
			name: "Empty",
			t:    time.Time{},
			want: "",
		},
		{
			name: "CET",
			t:    time.Date(2024, 3, 17, 10, 15, 0, 0, time.FixedZone("CET", 1*60*60)),
			tz:   "UTC",
			want: "Mar 17, 2024 UTC",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hd := humanDate(tc.tz, tc.t)
			assert.Equal(t, hd, tc.want)
		})
	}
}

func TestHumanDatetime(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		tz   string
		want string
	}{
		{
			name: "UTC",
			t:    time.Date(2024, 3, 17, 10, 15, 0, 0, time.UTC),
			tz:   "UTC",
			want: "Mar 17, 2024 10:15 AM UTC",
		},
		{
			name: "Empty",
			t:    time.Time{},
			want: "",
		},
		{
			name: "CET",
			t:    time.Date(2024, 3, 17, 10, 15, 0, 0, time.FixedZone("CET", 1*60*60)),
			tz:   "UTC",
			want: "Mar 17, 2024 09:15 AM UTC",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, humanDatetime(tc.tz, tc.t), tc.want)
		})
	}
}

func TestHumanTimeDiff(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{
			name: "Zero value",
			t:    time.Time{},
			want: "",
		},
		{
			name: "0 seconds ago",
			t:    time.Now(),
			want: "0 seconds ago",
		},
		{
			name: "1 second ago",
			t:    time.Now().Add(-1 * time.Second),
			want: "1 second ago",
		},
		{
			name: "1 minute ago",
			t:    time.Now().Add(-1 * time.Minute),
			want: "1 minute ago",
		},
		{
			name: "2 hours ago",
			t:    time.Now().Add(-2 * time.Hour),
			want: "2 hours ago",
		},
		{
			name: "3 days ago",
			t:    time.Now().Add(-24 * 3 * time.Hour),
			want: "3 days ago",
		},
		{
			name: "29 days ago",
			t:    time.Now().Add(-24 * 29 * time.Hour),
			want: "29 days ago",
		},
		{
			name: "1 month ago",
			t:    time.Now().Add(-24 * 30 * time.Hour),
			want: "1 month ago",
		},
		{
			name: "1 year ago",
			t:    time.Now().Add(-24 * 365 * time.Hour),
			want: "1 year ago",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, humanTimeDiff(tc.t), tc.want)
		})
	}
}
