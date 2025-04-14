package handlers_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/robyparr/event-horizon/internal/assert"
	"github.com/robyparr/event-horizon/internal/models"
)

func TestUserSignup(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app)
	defer ts.Close()

	validCSRFToken := ts.getCSRFToken(t)

	const (
		validPassword = "validPa$$word"
		validEmail    = "bob@example.com"
		formTag       = `<form action="/user/signup" method="POST" class="card" novalidate>`
	)

	app.Repos.Users.Insert(&models.User{ID: 1, Email: "dupe@example.com"})

	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantFormTag  string
	}{
		{
			name:         "Valid submission",
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusSeeOther,
		},
		{
			name:         "Invalid CSRF Token",
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    "wrongToken",
			wantCode:     http.StatusBadRequest,
		},
		{
			name:         "Empty email",
			userEmail:    "",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Empty password",
			userEmail:    validEmail,
			userPassword: "",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Invalid email",
			userEmail:    "bob@example.",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Short password",
			userEmail:    validEmail,
			userPassword: "pa$$",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
		{
			name:         "Duplicate email",
			userEmail:    "dupe@example.com",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := ts.postForm(t, "/user/signup", form)
			assert.Equal(t, code, tt.wantCode)

			if tt.wantFormTag != "" {
				assert.StringContains(t, body, tt.wantFormTag)
			}
		})
	}
}

func TestUserLogin(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app)
	defer ts.Close()

	validCSRFToken := ts.getCSRFToken(t)
	app.Repos.Users.Insert(&models.User{ID: 1, Email: "test@example.com", Password: "pa$$word"})

	tests := []struct {
		name         string
		email        string
		password     string
		csrfToken    string
		wantCode     int
		wantBody     string
		wantSignedIn bool
	}{
		{
			name:         "Valid credentials",
			email:        "test@example.com",
			password:     "pa$$word",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusSeeOther,
			wantSignedIn: true,
		},
		{
			name:      "Invalid email",
			email:     "invalid@example.com",
			password:  "pa$$word",
			csrfToken: validCSRFToken,
			wantCode:  http.StatusUnprocessableEntity,
			wantBody:  "Email or password is incorrect",
		},
		{
			name:      "Invalid password",
			email:     "test@example.com",
			password:  "invalid",
			csrfToken: validCSRFToken,
			wantCode:  http.StatusUnprocessableEntity,
			wantBody:  "Email or password is incorrect",
		},
		{
			name:      "Invalid CSRF token",
			email:     "test@example.com",
			password:  "pa$$word",
			csrfToken: "invalid",
			wantCode:  http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("csrf_token", tc.csrfToken)
			form.Add("email", tc.email)
			form.Add("password", tc.password)

			status, _, body := ts.postForm(t, "/user/login", form)
			assert.Equal(t, status, tc.wantCode)

			if tc.wantBody != "" {
				assert.StringContains(t, body, tc.wantBody)
			}

			if tc.wantSignedIn {
				cookie := ts.getCookie("session")
				assert.NotNil(t, cookie)
			} else {
				status, _, _ = ts.get(t, "/")
				assert.Equal(t, status, http.StatusSeeOther)
			}
		})
	}
}

func TestUserLogout(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app)
	defer ts.Close()

	ts.loginUser(t, models.User{ID: 1, Email: "test@example.com"})
	validCSRFToken := ts.getCSRFToken(t)

	tests := []struct {
		name               string
		csrfToken          string
		wantCode           int
		wantLocationHeader string
		wantLoggedOut      bool
	}{
		{
			name:               "Successful logout",
			csrfToken:          validCSRFToken,
			wantCode:           http.StatusSeeOther,
			wantLocationHeader: "/user/login",
			wantLoggedOut:      true,
		},
		{
			name:      "Invalid CSRF token",
			csrfToken: "invalid",
			wantCode:  http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("csrf_token", tc.csrfToken)

			status, headers, _ := ts.postForm(t, "/user/logout", form)
			assert.Equal(t, status, tc.wantCode)
			assert.Equal(t, headers.Get("Location"), tc.wantLocationHeader)

			cookie := ts.getCookie("session")
			if tc.wantLoggedOut {
				assert.Equal(t, cookie, nil)
			} else {
				assert.NotNil(t, cookie)
			}
		})
	}
}

func TestUserSettings(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app)
	defer ts.Close()

	ts.loginUser(t, models.User{ID: 1, Email: "test@example.com"})

	status, _, body := ts.get(t, "/user/settings")
	assert.Equal(t, status, http.StatusOK)
	assert.StringContains(t, body, "Sessions")
	assert.StringContains(t, body, "Current")
	assert.StringContains(t, body, "/sessions/1/delete")
}

func TestUserSessionsDelete(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app)
	defer ts.Close()

	ts.loginUser(t, models.User{ID: 1, Email: "test@example.com"})
	csrfToken := ts.getCSRFToken(t)

	status, _, _ := ts.get(t, "/user/settings")
	assert.Equal(t, status, http.StatusOK)

	form := url.Values{}
	form.Add("csrf_token", csrfToken)
	status, _, _ = ts.postForm(t, "/sessions/1/delete", form)
	assert.Equal(t, status, http.StatusSeeOther)
}
