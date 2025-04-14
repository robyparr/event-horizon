package views

import (
	"fmt"
	"html/template"
	"math"
	"time"
)

const dateFormat = "Jan 02, 2006"
const timeFormat = "03:04 PM"

var functions = template.FuncMap{
	"humanDate":     humanDate,
	"humanDatetime": humanDatetime,
	"humanTimeDiff": humanTimeDiff,
}

func humanDate(tz string, t time.Time) string {
	if t.IsZero() {
		return ""
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return t.UTC().Format(dateFormat + " MST")
	}

	return t.In(loc).Format(dateFormat + " MST")
}

func humanDatetime(tz string, t time.Time) string {
	if t.IsZero() {
		return ""
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return t.UTC().Format(dateFormat + " MST")
	}

	return t.In(loc).Format(dateFormat + " " + timeFormat + " MST")
}

func humanTimeDiff(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	dur := time.Now().UTC().Sub(t.UTC()).Truncate(time.Second)
	relativeWord := "ago"
	if dur.Seconds() < 0 {
		dur = -dur
		relativeWord = "from now"
	}

	var v float64
	var timeUnit string
	switch {
	case dur.Seconds() < 60:
		v = dur.Seconds()
		timeUnit = "second"
	case dur.Minutes() < 60:
		v = dur.Minutes()
		timeUnit = "minute"
	case dur.Hours() < 24:
		v = dur.Hours()
		timeUnit = "hour"
	case dur.Hours()/24 < 30:
		v = dur.Hours() / 24
		timeUnit = "day"
	case dur.Hours()/24 < 365:
		v = dur.Hours() / 24 / 30
		timeUnit = "month"
	default:
		v = dur.Hours() / 24 / 30 / 12
		timeUnit = "year"
	}

	var vInt int = int(math.Round(v))
	return fmt.Sprintf("%d %s %s", vInt, pluralize(timeUnit, vInt), relativeWord)
}

func pluralize(word string, n int) string {
	if n == 1 {
		return word
	}

	return word + "s"
}
