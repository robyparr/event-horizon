package internal

import (
	"context"
	"net/http"

	"github.com/robyparr/event-horizon/internal/models"
)

type ctxKey string

const ctxKeySession = ctxKey("session")
const ctxKeyCurrentUser = ctxKey("currentUser")
const ctxKeyCurrentSite = ctxKey("currentSite")
const ctxKeyFlash = ctxKey("flash")

func (app *App) SetFlashDataInContext(r *http.Request, flashData map[string]string) *http.Request {
	ctx := context.WithValue(r.Context(), ctxKeyFlash, flashData)
	return r.WithContext(ctx)
}

func (app *App) GetFlashDataFromContext(r *http.Request) map[string]string {
	data, _ := r.Context().Value(ctxKeyFlash).(map[string]string)
	return data
}

func (app *App) SetCurrentSite(r *http.Request, site models.Site) *http.Request {
	ctx := context.WithValue(r.Context(), ctxKeyCurrentSite, site)
	return r.WithContext(ctx)
}

func (app *App) MustGetCurrentSite(r *http.Request) *models.Site {
	site, ok := r.Context().Value(ctxKeyCurrentSite).(models.Site)
	if !ok || site.ID == 0 {
		panic("Current site must exist")
	}

	return &site
}
