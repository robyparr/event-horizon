package assert

import (
	"strings"
	"testing"
)

func Equal[T comparable](t *testing.T, actual T, expected T) {
	t.Helper()

	if actual != expected {
		t.Errorf("got: %v, want: %v", actual, expected)
	}
}

func NotEqual[T comparable](t *testing.T, actual T, expected T) {
	t.Helper()

	if actual == expected {
		t.Errorf("expected a and b not to be equal: %v", actual)
	}
}

func StringContains(t *testing.T, actual string, expectedSubstring string) {
	t.Helper()

	if !strings.Contains(actual, expectedSubstring) {
		t.Errorf("got: %q, expected to contain: %q", actual, expectedSubstring)
	}
}

func Nil(t *testing.T, actual any) {
	t.Helper()

	if actual != nil {
		t.Errorf("got: %q, expected: nil", actual)
	}
}

func NotNil(t *testing.T, actual any) {
	t.Helper()

	if actual == nil {
		t.Errorf("got: nil, expected: not nil")
	}
}
