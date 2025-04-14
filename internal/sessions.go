package internal

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"github.com/robyparr/event-horizon/internal/models"
	"github.com/tomasen/realip"
)

func (app *App) SetSession(r *http.Request, session models.Session) *http.Request {
	ctx := context.WithValue(r.Context(), ctxKeySession, session)
	ctx = context.WithValue(ctx, ctxKeyCurrentUser, session.User)
	return r.WithContext(ctx)
}

func (app *App) GetSession(r *http.Request) (models.Session, bool) {
	s, ok := r.Context().Value(ctxKeySession).(models.Session)
	return s, ok
}

func (app *App) MustGetSession(r *http.Request) models.Session {
	s, ok := app.GetSession(r)
	if !ok || s.ID <= 0 {
		panic("No session available")
	}

	return s
}

func (app *App) GetCurrentUser(r *http.Request) (models.User, bool) {
	u, ok := r.Context().Value(ctxKeyCurrentUser).(models.User)
	return u, ok
}

func (app *App) MustGetCurrentUser(r *http.Request) models.User {
	u, ok := app.GetCurrentUser(r)
	if !ok || u.ID <= 0 {
		panic("No current user available")
	}

	return u
}

func (app *App) CreateSession(w http.ResponseWriter, r *http.Request, userID int64) error {
	token := rand.Text()
	s := models.Session{
		UserID:    userID,
		Token:     fmt.Sprintf("%x", sha256.Sum256([]byte(token))),
		IPAddress: realip.FromRequest(r),
		UserAgent: r.UserAgent(),
		ExpiresAt: time.Now().Add(24 * time.Hour * 30).UTC(),
	}

	err := app.Repos.Sessions.Insert(&s)
	if err != nil {
		return fmt.Errorf("[CreateSession] %w", err)
	}

	cookie, err := app.BuildSignedCookie("session", map[string]string{"token": token})
	if err != nil {
		return fmt.Errorf("[CreateSession] %w", err)
	}

	cookie.MaxAge = int(s.ExpiresAt.Sub(time.Now().UTC()).Seconds())
	http.SetCookie(w, cookie)
	return nil
}

func (app *App) DeleteSession(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	session, ok := app.GetSession(r)
	if !ok {
		panic("[DeleteSession] no session to delete")
	}

	err := app.Repos.Sessions.Delete(&session)
	if err != nil {
		return r, fmt.Errorf("[DeleteSession] %w", err)
	}

	ctx := context.WithValue(r.Context(), ctxKeySession, nil)
	ctx = context.WithValue(ctx, ctxKeyCurrentUser, nil)
	r = r.WithContext(ctx)

	cookie, err := app.BuildSignedCookie("session", map[string]string{"token": ""})
	if err != nil {
		return r, fmt.Errorf("[DeleteSession] %w", err)
	}

	http.SetCookie(w, cookie)
	return r, nil
}
