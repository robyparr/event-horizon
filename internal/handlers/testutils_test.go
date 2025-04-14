package handlers_test

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/robyparr/event-horizon/internal"
	"github.com/robyparr/event-horizon/internal/assert"
	"github.com/robyparr/event-horizon/internal/handlers"
	"github.com/robyparr/event-horizon/internal/models"
	"github.com/robyparr/event-horizon/internal/models/mocks"
	"github.com/robyparr/event-horizon/internal/views"

	"github.com/go-playground/form/v4"
)

var csrfTokenRx = regexp.MustCompile(`<input type="hidden" name="csrf_token" value="(.+)" />`)

func newTestApplication(t *testing.T) *internal.App {
	views, err := views.CompileViews()
	if err != nil {
		t.Fatal(err)
	}

	return &internal.App{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Repos: &models.Repos{
			Users:    mocks.NewUserRepo(),
			Sessions: mocks.NewSessionRepo(),
		},
		Views:        views,
		FormDecoder:  form.NewDecoder(),
		SecureCookie: securecookie.New([]byte("super secret"), nil),
	}
}

type testServer struct {
	*httptest.Server
	app *internal.App
}

func newTestServer(t *testing.T, app *internal.App) *testServer {
	ts := httptest.NewTLSServer(handlers.Routes(app))

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	ts.Client().Jar = jar
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{
		Server: ts,
		app:    app,
	}
}

func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	result, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatal(err)
	}

	body = bytes.TrimSpace(body)
	return result.StatusCode, result.Header, string(body)
}

func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) (int, http.Header, string) {
	result, err := ts.Client().PostForm(ts.URL+urlPath, form)
	if err != nil {
		t.Fatal(err)
	}

	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatal(err)
	}

	body = bytes.TrimSpace(body)
	return result.StatusCode, result.Header, string(body)
}

func (ts *testServer) loginUser(t *testing.T, user models.User) {
	token := rand.Text()
	session := models.Session{User: user, UserID: user.ID, Token: fmt.Sprintf("%x", sha256.Sum256([]byte(token)))}
	assert.Nil(t, ts.app.Repos.Sessions.Insert(&session))

	cookie, err := ts.app.BuildSignedCookie("session", map[string]string{"token": token})
	assert.Nil(t, err)

	cookie.MaxAge = 1
	url, _ := url.Parse(ts.URL)
	ts.Client().Jar.SetCookies(url, []*http.Cookie{cookie})
}

func (ts *testServer) getCookie(name string) *http.Cookie {
	url, _ := url.Parse(ts.URL)
	for _, cookie := range ts.Client().Jar.Cookies(url) {
		if cookie.Name == name {
			return cookie
		}
	}

	return nil
}

func (ts *testServer) getCSRFToken(t *testing.T) string {
	t.Helper()

	_, _, body := ts.get(t, "/user/login")
	matches := csrfTokenRx.FindStringSubmatch(body)
	if len(matches) < 2 {
		t.Fatalf("no CSRF token found in body: %s", body)
	}

	return html.UnescapeString(matches[1])
}
