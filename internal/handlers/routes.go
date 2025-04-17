package handlers

import (
	"net/http"

	"github.com/justinas/alice"
	"github.com/robyparr/event-horizon/internal"
	"github.com/robyparr/event-horizon/internal/views"
)

func Routes(app *internal.App) http.Handler {
	mux := http.NewServeMux()
	middleware := middleware{app: app}

	mux.Handle("GET /static/", middleware.cacheStaticAssets(http.FileServerFS(views.Files)))
	mux.HandleFunc("GET /healthcheck", healthcheckHandler)

	standardMiddleware := alice.New(middleware.handleSession, middleware.noSurf, middleware.flash)

	// User Authentication
	mux.Handle("GET /user/signup", standardMiddleware.Then(userSignupFormHandler(app)))
	mux.Handle("POST /user/signup", standardMiddleware.Then(userSignupHandler(app)))
	mux.Handle("GET /user/login", standardMiddleware.Then(userLoginFormHandler(app)))
	mux.Handle("POST /user/login", standardMiddleware.Then(userLoginHandler(app)))

	// Authentication required
	requireAuthMiddleware := standardMiddleware.Append(middleware.requireAuthentication)
	mux.Handle("GET /{$}", requireAuthMiddleware.Then(sitesListHandler(app)))

	mux.Handle("POST /user/logout", requireAuthMiddleware.Then(userLogoutPostHandler(app)))
	mux.Handle("GET /user/settings", requireAuthMiddleware.Then(userSettingsHandler(app)))
	mux.Handle("POST /sessions/{id}/delete", requireAuthMiddleware.Then(userSessionsDeleteHandler(app)))

	mux.Handle("GET /sites/{id}", requireAuthMiddleware.Then(sitesShowHandler(app)))
	mux.Handle("POST /sites", requireAuthMiddleware.Then(sitesCreateHandler(app)))
	mux.Handle("POST /sites/{id}/delete", requireAuthMiddleware.Then(sitesDeleteHandler(app)))

	baseMiddleware := alice.New(middleware.recoverPanic, middleware.logRequest, middleware.commonHeaders)
	return baseMiddleware.Then(mux)
}
