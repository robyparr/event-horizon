package models

import (
	"testing"

	"github.com/robyparr/event-horizon/internal/assert"
)

func TestUserInsert(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

	const email = "test@example.com"
	const password = "pa$$word"

	t.Run("Insert a user", func(t *testing.T) {
		r := UserRepo{db: newTestDB(t)}
		id, _ := r.Authenticate(email, password)
		assert.Equal(t, id, 0)

		user := User{Email: email, Password: password}
		assert.Nil(t, r.Insert(&user))

		id, _ = r.Authenticate(email, password)
		assert.Equal(t, id, user.ID)
		if user.ID <= 0 {
			t.Errorf("got: %q, want: a positive integer", user.ID)
		}
	})

	t.Run("Duplicate email", func(t *testing.T) {
		r := UserRepo{db: newTestDB(t)}
		r.Insert(&User{Email: email, Password: password})

		user := User{Email: email, Password: password}
		err := r.Insert(&user)
		assert.NotNil(t, err)

		if err != ErrDuplicateEmail {
			t.Errorf("got: %q, want: ErrDuplicateEmail", err)
		}

		assert.Equal(t, user.ID, 0)
	})
}

func TestUserAuthenticate(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

	r := UserRepo{db: newTestDB(t)}

	user1 := &User{Email: "one@example.com", Password: "pa$$word"}
	assert.Nil(t, r.Insert(user1))

	user2 := &User{Email: "two@example.com", Password: "pa$$word"}
	assert.Nil(t, r.Insert(user2))

	t.Run("successful authentication", func(t *testing.T) {
		id, err := r.Authenticate(user2.Email, user2.Password)
		assert.Nil(t, err)
		assert.Equal(t, id, user2.ID)
	})

	t.Run("Failed authentication", func(t *testing.T) {
		id, err := r.Authenticate(user2.Email, "wrong password")
		assert.Equal(t, err, ErrInvalidCredentials)
		assert.Equal(t, id, 0)
	})

	t.Run("Invalid user", func(t *testing.T) {
		id, err := r.Authenticate("wrong@example.com", "pa$$word")
		assert.Equal(t, err, ErrInvalidCredentials)
		assert.Equal(t, id, 0)
	})
}
