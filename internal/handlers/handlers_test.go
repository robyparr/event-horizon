package handlers_test

import (
	"net/http"
	"testing"

	"github.com/robyparr/event-horizon/internal/assert"
	"github.com/robyparr/event-horizon/internal/models"
)

func TestHomeHandler(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app)
	defer ts.Close()

	t.Run("Not logged in", func(t *testing.T) {
		code, _, body := ts.get(t, "/")
		assert.Equal(t, code, http.StatusSeeOther)
		assert.StringContains(t, body, `<a href="/user/login">See Other</a>.`)
	})

	t.Run("Logged in", func(t *testing.T) {
		user := models.User{ID: 1, Email: "test@example.com"}
		ts.loginUser(t, user)

		code, _, body := ts.get(t, "/")
		assert.Equal(t, code, http.StatusOK)
		assert.StringContains(t, body, "<h1>Home</h1>")
		assert.StringContains(t, body, user.Email)
	})
}

func TestHealthcheck(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app)
	defer ts.Close()

	code, _, body := ts.get(t, "/healthcheck")
	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, body, "OK")
}
