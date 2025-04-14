package internal

import (
	"context"
	"net/http"
)

type ctxKey string

const ctxKeySession = ctxKey("session")
const ctxKeyCurrentUser = ctxKey("currentUser")
const ctxKeyFlash = ctxKey("flash")

func (app *App) SetFlashDataInContext(r *http.Request, flashData map[string]string) *http.Request {
	ctx := context.WithValue(r.Context(), ctxKeyFlash, flashData)
	return r.WithContext(ctx)
}

func (app *App) GetFlashDataFromContext(r *http.Request) map[string]string {
	data, _ := r.Context().Value(ctxKeyFlash).(map[string]string)
	return data
}
