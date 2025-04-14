package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/robyparr/event-horizon/internal"
	"github.com/robyparr/event-horizon/internal/models"
	"github.com/robyparr/event-horizon/internal/validator"
	"github.com/robyparr/event-horizon/internal/views"
)

type userSignupForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	Timezone            string `form:"timezone"`
	validator.Validator `form:"-"`
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func userSignupFormHandler(app *internal.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.Render(w, r, http.StatusOK, "user/signup.html.tmpl", views.NewViewModel(app, r, userSignupForm{}))
	})
}

func userSignupHandler(app *internal.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var form userSignupForm
		err := app.DecodePostForm(r, &form)
		if err != nil {
			app.ClientError(w, http.StatusBadRequest)
			return
		}

		form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
		form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
		form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
		form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

		if !form.Valid() {
			app.Render(w, r, http.StatusUnprocessableEntity, "user/signup.html.tmpl", views.NewViewModel(app, r, form))
			return
		}

		user := models.User{Email: form.Email, Password: form.Password, Timezone: form.Timezone}
		err = app.Repos.Users.Insert(&user)
		if err != nil {
			if errors.Is(err, models.ErrDuplicateEmail) {
				form.AddFieldError("email", "Email address is already in use")
				app.Render(w, r, http.StatusUnprocessableEntity, "user/signup.html.tmpl", views.NewViewModel(app, r, form))
			} else {
				app.ServerError(w, r, err)
			}

			return
		}

		err = app.CreateSession(w, r, user.ID)
		if err != nil {
			app.ServerError(w, r, err)
			return
		}

		app.SetFlash(w, "info", "Welcome!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}

func userLoginFormHandler(app *internal.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.Render(w, r, http.StatusOK, "user/login.html.tmpl", views.NewViewModel(app, r, userLoginForm{}))
	})
}

func userLoginHandler(app *internal.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var form userLoginForm

		err := app.DecodePostForm(r, &form)
		if err != nil {
			app.ClientError(w, http.StatusBadRequest)
			return
		}

		form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
		form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
		form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

		if !form.Valid() {
			app.Render(w, r, http.StatusUnprocessableEntity, "user/login.html.tmpl", views.NewViewModel(app, r, form))
			return
		}

		id, err := app.Repos.Users.Authenticate(form.Email, form.Password)
		if err != nil {
			if errors.Is(err, models.ErrInvalidCredentials) {
				form.AddNonFieldError("Email or password is incorrect")
				app.Render(w, r, http.StatusUnprocessableEntity, "user/login.html.tmpl", views.NewViewModel(app, r, form))
			} else {
				app.ServerError(w, r, err)
			}

			return
		}

		err = app.CreateSession(w, r, id)
		if err != nil {
			app.ServerError(w, r, err)
			return
		}

		username := strings.Split(form.Email, "@")[0]
		app.SetFlash(w, "info", fmt.Sprintf("Hi, %s!", username))
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}

func userLogoutPostHandler(app *internal.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r, err := app.DeleteSession(w, r)
		if err != nil {
			app.ServerError(w, r, err)
			return
		}

		app.SetFlash(w, "info", "You've been logged out successfully!")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
	})
}

func userSettingsHandler(app *internal.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentUser := app.MustGetCurrentUser(r)
		sessions, err := app.Repos.Sessions.ListForUser(&currentUser)
		if err != nil {
			app.ServerError(w, r, err)
			return
		}

		currentSession := app.MustGetSession(r)
		for i, session := range sessions {
			if session.Token == currentSession.Token {
				sessions[i].Current = true
			}
		}

		vm := views.NewViewModel(app, r, nil)
		vm.Data["sessions"] = sessions
		app.Render(w, r, http.StatusOK, "user/settings.html.tmpl", vm)
	})
}

func userSessionsDeleteHandler(app *internal.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := readIDParam(r)
		if err != nil {
			app.ServerError(w, r, err)
			return
		}

		currentUser := app.MustGetCurrentUser(r)
		err = app.Repos.Sessions.DeleteByID(currentUser, id)
		if err != nil {
			app.ServerError(w, r, err)
			return
		}

		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
	})
}
