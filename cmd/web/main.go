package main

import (
	"cmp"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/gorilla/securecookie"
	"github.com/robyparr/event-horizon/internal"
	"github.com/robyparr/event-horizon/internal/handlers"
	"github.com/robyparr/event-horizon/internal/models"
	"github.com/robyparr/event-horizon/internal/views"

	_ "github.com/lib/pq"
)

func main() {
	config := internal.Config{
		Environment: cmp.Or(os.Getenv("ENV"), internal.EnvDevelopment),
		Host:        cmp.Or(os.Getenv("HOST"), ":4000"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		SkipTLS:     os.Getenv("SKIP_TLS") == "true",
	}

	views, err := views.CompileViews()
	if err != nil {
		log.Fatalf("[CompileViews] %s", err.Error())
	}

	db, err := openDB(config.DatabaseURL)
	if err != nil {
		log.Fatalf("[openDB] %s", err.Error())
	}
	defer db.Close()

	cookieSecretKey := cmp.Or(
		os.Getenv("COOKIE_SECRET_KEY"),
		base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(64)),
	)
	app := &internal.App{
		Logger:       slog.New(slog.NewTextHandler(os.Stdout, nil)),
		FormDecoder:  form.NewDecoder(),
		Config:       config,
		Views:        views,
		Repos:        models.NewRepos(db),
		SecureCookie: securecookie.New([]byte(cookieSecretKey), nil),
		StartedAt:    time.Now().UTC(),
	}

	server := &http.Server{
		Addr:         app.Config.Host,
		Handler:      handlers.Routes(app),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.Logger.Handler(), slog.LevelError),
		TLSConfig: &tls.Config{
			CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		},
	}

	startBackgroundProcesses(app)
	shutdownError := gracefulShutdown(app, server)
	app.Logger.Info("starting server", "address", server.Addr)
	if app.Config.SkipTLS {
		err = server.ListenAndServe()
	} else {
		err = server.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	}

	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("[ListenAndServe] %s", err.Error())
	}

	err = <-shutdownError
	if err != nil {
		log.Fatalf("[shutdownError] %s", err.Error())
	}

	app.Logger.Info("Stopped server", "address", server.Addr)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxIdleTime(15 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func gracefulShutdown(app *internal.App, server *http.Server) chan error {
	shutdownError := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit
		app.Logger.Info("caught signal", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := server.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.Logger.Info("Waiting for background tasks...", "address", server.Addr)

		app.WG.Wait()
		shutdownError <- nil
	}()

	return shutdownError
}

func startBackgroundProcesses(app *internal.App) {
	app.StartProcess("session cleanup", func() []any {
		for {
			time.Sleep(24 * time.Hour)

			app.InBackground("session cleanup", func() []any {
				count, err := app.Repos.Sessions.DeleteExpired()
				if err != nil {
					app.Logger.Error("Unable to clear sessions", "error", err.Error())
					return []any{}
				}

				return []any{"count", count}
			})
		}
	})
}
