package handlers

import (
	"cmp"
	"encoding/json"
	"net/http"

	"github.com/mileusna/useragent"
	"github.com/robyparr/event-horizon/internal"
	"github.com/robyparr/event-horizon/internal/models"
)

func apiPreflightHandler(_ *internal.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMethod := r.Header.Get("Access-Control-Request-Method")
		if requestMethod == "" {
			http.NotFound(w, r)
			return
		}

		if requestMethod != "POST" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}

func apiCreateEventHandler(app *internal.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := useragent.Parse(r.UserAgent())
		if ua.Bot {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		var eventData struct {
			Action string `json:"action"`
			Count  int    `json:"count"`
		}
		err := json.NewDecoder(r.Body).Decode(&eventData)
		if err != nil {
			app.ServerError(w, r, err)
			return
		}

		site := app.MustGetCurrentSite(r)
		event := models.Event{
			SiteID:     site.ID,
			Action:     eventData.Action,
			Count:      eventData.Count,
			DeviceType: cmp.Or(ua.Device, "Unknown"),
			OS:         cmp.Or(ua.OS, "Unknown"),
			Browser:    cmp.Or(ua.Name, "Unknown"),
		}

		err = app.Repos.Events.Insert(&event)
		if err != nil {
			app.ServerError(w, r, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}
