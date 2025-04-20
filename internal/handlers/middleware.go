package handlers

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/justinas/nosurf"
	"github.com/robyparr/event-horizon/internal"
	"github.com/robyparr/event-horizon/internal/models"
	"github.com/robyparr/event-horizon/internal/utils"
)

var skipLoggingURLPrefixes = []string{
	"/static/",
	"/apple-touch-icon",
	"favicon.ico",
}

type middleware struct {
	app *internal.App
}

func (m middleware) commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scriptNonce := utils.Token()
		w.Header().Set("Content-Security-Policy", fmt.Sprintf("default-src 'self'; style-src 'self'; script-src 'self' 'nonce-%s'", scriptNonce))
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		r = m.app.SetScriptNonce(r, scriptNonce)
		next.ServeHTTP(w, r)
	})
}

func (m middleware) commonAPIHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Access-Control-Request-Method")

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		next.ServeHTTP(w, r)
	})
}

func (m middleware) loadSite(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token, found := strings.CutPrefix(authHeader, "Bearer ")
		if !found {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		site, err := m.app.Repos.Sites.FindByToken(token)
		if err != nil {
			if errors.Is(err, models.ErrNoRecord) {
				http.NotFound(w, r)
				return
			}

			m.app.ServerError(w, r, err)
			return
		}

		r = m.app.SetCurrentSite(r, site)
		next.ServeHTTP(w, r)
	})
}

func (m middleware) cacheStaticAssets(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.app.Config.IsProductionEnv() {
			w.Header().Set("Cache-Control", "public, max-age=86400, must-revalidate")
		}

		next.ServeHTTP(w, r)
	})
}

func (m middleware) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.app.Config.IsProductionEnv() && hasSkippableURLPrefix(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		next.ServeHTTP(w, r)
		dur := time.Since(start)

		m.app.Logger.Info("received request", "ip", r.RemoteAddr, "proto", r.Proto, "method", r.Method, "uri", r.URL.RequestURI(), "time", dur.String())
	})
}

func hasSkippableURLPrefix(url string) bool {
	for _, prefix := range skipLoggingURLPrefixes {
		if strings.HasPrefix(url, prefix) {
			return true
		}
	}

	return false
}

func (m middleware) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				m.app.ServerError(w, r, fmt.Errorf("[recoverPanic] %s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (m middleware) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.app.IsAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func (m middleware) noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}

func (m middleware) handleSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieData, err := m.app.ReadSignedCookie(w, r, "session")
		if err != nil {
			m.app.ServerError(w, r, fmt.Errorf("[HandleSession] %w", err))
			return
		}

		if cookieData["token"] == "" {
			next.ServeHTTP(w, r)
			return
		}

		hashedToken := fmt.Sprintf("%x", sha256.Sum256([]byte(cookieData["token"])))
		session, err := m.app.Repos.Sessions.FindByToken(hashedToken)
		if err != nil {
			if errors.Is(err, models.ErrNoRecord) {
				next.ServeHTTP(w, r)
				return
			}

			m.app.ServerError(w, r, fmt.Errorf("[HandleSession] %w", err))
			return
		}

		r = m.app.SetSession(r, session)
		next.ServeHTTP(w, r)

		err = m.app.Repos.Sessions.Touch(&session)
		if err != nil {
			panic(fmt.Sprintf("[handleSession] %s", err.Error()))
		}
	})
}

func (m middleware) flash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flashData, err := m.app.ReadSignedCookie(w, r, "flash")
		if err != nil {
			m.app.ServerError(w, r, err)
			return
		}

		cookie, _ := m.app.BuildSignedCookie("flash", map[string]string{})
		http.SetCookie(w, cookie)

		r = m.app.SetFlashDataInContext(r, flashData)
		next.ServeHTTP(w, r)
	})
}
