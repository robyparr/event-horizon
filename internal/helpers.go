package internal

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/form/v4"
)

func (app *App) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	method := r.Method
	uri := r.URL.RequestURI()

	app.Logger.Error(err.Error(), slog.String("method", method), slog.String("uri", uri))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *App) ClientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *App) Render(w http.ResponseWriter, r *http.Request, status int, page string, data any) {
	ts, ok := app.Views[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.ServerError(w, r, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (app *App) DecodePostForm(r *http.Request, dest any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.FormDecoder.Decode(dest, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		return err
	}

	return nil
}

func (app *App) IsAuthenticated(r *http.Request) bool {
	user, ok := app.GetCurrentUser(r)
	if !ok {
		return false
	}

	return user.ID > 0
}

func (app *App) BuildSignedCookie(name string, values map[string]string) (*http.Cookie, error) {
	encodedValue, err := app.SecureCookie.Encode(name, values)
	if err != nil {
		return nil, fmt.Errorf("[BuildSignedCookie] %w", err)
	}

	return &http.Cookie{
		Name:     name,
		Value:    encodedValue,
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}, nil

}

func (app *App) ReadSignedCookie(w http.ResponseWriter, r *http.Request, name string) (map[string]string, error) {
	data := make(map[string]string)
	cookie, err := r.Cookie(name)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return data, nil
		}

		return data, fmt.Errorf("[ReadSignedCookie] %w", err)
	}

	err = app.SecureCookie.Decode(name, cookie.Value, &data)
	if err != nil {
		return data, fmt.Errorf("[ReadSignedCookie] %w", err)
	}

	return data, nil
}

func (app *App) SetFlash(w http.ResponseWriter, flashType string, msg string) {
	values := make(map[string]string)
	values[flashType] = msg

	cookie, err := app.BuildSignedCookie("flash", values)
	if err != nil {
		return
	}

	cookie.MaxAge = 0
	http.SetCookie(w, cookie)
}

func (app *App) InBackground(name string, fn func() []any) {
	app.runInBackground(name, false, fn)
}

func (app *App) StartProcess(name string, fn func() []any) {
	app.runInBackground(name, true, fn)
}

func (app *App) runInBackground(name string, persistent bool, fn func() []any) {
	if !persistent {
		app.WG.Add(1)
	}

	go func() {
		if !persistent {
			defer app.WG.Done()
		}
		start := time.Now()

		defer func() {
			if err := recover(); err != nil {
				app.Logger.Error(fmt.Sprintf("%v", err))
			}
		}()

		logArgs := []any{"name", name, "time", time.Since(start).Truncate(time.Microsecond)}
		logArgs = append(logArgs, fn()...)
		app.Logger.Info("Ran background task", logArgs...)
	}()
}
