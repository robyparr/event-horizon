package handlers

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robyparr/event-horizon/internal/assert"
)

func TestCommonHeaders(t *testing.T) {
	rr := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	middleware{}.commonHeaders(next).ServeHTTP(rr, r)
	result := rr.Result()

	expected := "default-src 'self'; style-src 'self';"
	assert.Equal(t, result.Header.Get("Content-Security-Policy"), expected)

	assert.Equal(t, result.Header.Get("Referrer-Policy"), "origin-when-cross-origin")
	assert.Equal(t, result.Header.Get("X-Content-Type-Options"), "nosniff")
	assert.Equal(t, result.Header.Get("X-XSS-Protection"), "0")

	assert.Equal(t, result.StatusCode, http.StatusOK)

	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatal(err)
	}

	body = bytes.TrimSpace(body)
	assert.Equal(t, string(body), "OK")
}
