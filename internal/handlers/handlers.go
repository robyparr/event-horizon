package handlers

import (
	"net/http"

	"github.com/robyparr/event-horizon/internal"
	"github.com/robyparr/event-horizon/internal/views"
)

func homeHandler(app *internal.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.Render(w, r, http.StatusOK, "home.html.tmpl", views.NewViewModel(app, r, nil))
	})
}

func healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
