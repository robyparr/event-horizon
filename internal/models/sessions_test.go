package models

import (
	"database/sql"
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/robyparr/event-horizon/internal/assert"
)

func buildDefaultSession(user *User) *Session {
	return &Session{
		UserID:    user.ID,
		Token:     fmt.Sprintf("%d", rand.Uint64()),
		IPAddress: "127.0.0.1",
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/603.3.8 (KHTML, like Gecko) Version/10.1.2 Safari/603.3.8",
		ExpiresAt: time.Now().Add(1 * time.Hour).In(time.UTC),
		User:      *user,
	}
}

func setupSession(t *testing.T) (*Session, *SessionRepo) {
	db := newTestDB(t)
	return setupSessionWithDB(t, db)
}

func setupSessionWithDB(t *testing.T, db *sql.DB) (*Session, *SessionRepo) {
	userRepo := UserRepo{db: db}
	sessionRepo := SessionRepo{db: db}

	user := &User{Email: fmt.Sprintf("%d%s", rand.Uint64(), "@example.com"), Password: "pa$$word"}
	assert.Nil(t, userRepo.Insert(user))
	session := buildDefaultSession(user)

	assert.Nil(t, sessionRepo.Insert(session))
	return session, &sessionRepo
}

func TestSessionInsert(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

	session, sessionRepo := setupSession(t)
	assert.Equal(t, session.ID, 1)

	t.Run("Inserted the session correctly", func(t *testing.T) {
		sessFromDB, err := sessionRepo.FindByToken(session.Token)
		assert.Nil(t, err)

		assert.Equal(t, sessFromDB.ID, session.ID)
		assert.Equal(t, sessFromDB.IPAddress, session.IPAddress)
		assert.Equal(t, sessFromDB.UserAgent, session.UserAgent)
		assert.Equal(t, sessFromDB.ExpiresAt.Local(), session.ExpiresAt.Local())
	})

	t.Run("Duplicate token", func(t *testing.T) {
		err := sessionRepo.Insert(&Session{UserID: session.UserID, Token: session.Token})
		assert.Equal(t, err, ErrDuplicateToken)
	})
}

func TestSessionDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

	session, sessionRepo := setupSession(t)

	assert.Nil(t, sessionRepo.Delete(session))
	_, err := sessionRepo.FindByToken(session.Token)
	assert.Equal(t, err, ErrNoRecord)
}

func TestSessionFindByToken(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

	session, sessionRepo := setupSession(t)
	assert.Equal(t, session.ID, 1)

	t.Run("finds a session", func(t *testing.T) {
		sessFromDB, err := sessionRepo.FindByToken(session.Token)
		assert.Nil(t, err)
		assert.Equal(t, sessFromDB.ID, session.ID)
		assert.Equal(t, sessFromDB.UserID, session.UserID)
		assert.Equal(t, sessFromDB.IPAddress, session.IPAddress)
		assert.Equal(t, sessFromDB.UserAgent, session.UserAgent)

		assert.Equal(t, sessFromDB.User.ID, session.User.ID)
		assert.Equal(t, sessFromDB.User.Email, session.User.Email)
	})

	t.Run("expired session", func(t *testing.T) {
		expiredSession := buildDefaultSession(&session.User)
		expiredSession.Token = "expired-token"
		expiredSession.ExpiresAt = time.Now().Add(-1 * time.Second).UTC()
		assert.Nil(t, sessionRepo.Insert(expiredSession))

		_, err := sessionRepo.FindByToken("expired-token")
		assert.Equal(t, err, ErrNoRecord)
	})

	t.Run("no such token", func(t *testing.T) {
		_, err := sessionRepo.FindByToken("no-such-token")
		assert.Equal(t, err, ErrNoRecord)
	})
}

func TestDeleteByID(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

	t.Run("Deleting a user session", func(t *testing.T) {
		session, repo := setupSession(t)
		sessionFromDB, err := repo.FindByToken(session.Token)
		assert.Nil(t, err)

		assert.Equal(t, session.ID, sessionFromDB.ID)
		err = repo.DeleteByID(session.User, session.ID)
		assert.Nil(t, err)

		_, err = repo.FindByToken(session.Token)
		assert.Equal(t, err, ErrNoRecord)
	})

	t.Run("Attempting to delete another user's session", func(t *testing.T) {
		db := newTestDB(t)
		session, repo := setupSessionWithDB(t, db)
		session2, _ := setupSessionWithDB(t, db)
		assert.NotEqual(t, session.UserID, session2.UserID)

		err := repo.DeleteByID(session2.User, session.ID)
		assert.Equal(t, err, ErrNoRecord)

		sessionFromDB, err := repo.FindByToken(session.Token)
		assert.Nil(t, err)
		assert.Equal(t, sessionFromDB.ID, session.ID)

		sessionFromDB, err = repo.FindByToken(session2.Token)
		assert.Nil(t, err)
		assert.Equal(t, sessionFromDB.ID, session2.ID)
	})
}

func TestListForUser(t *testing.T) {
	session, repo := setupSession(t)
	session2 := buildDefaultSession(&session.User)
	session2.UserAgent = "Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_2 like Mac OS X) AppleWebKit/603.2.4 (KHTML, like Gecko) Version/10.0 Mobile/14F89 Safari/602.1"
	err := repo.Insert(session2)
	assert.Nil(t, err)

	t.Run("List sessions", func(t *testing.T) {
		sessions, err := repo.ListForUser(&session.User)
		assert.Nil(t, err)

		assert.Equal(t, len(sessions), 2)
		assert.Equal(t, sessions[0].ID, session2.ID)
		assert.Equal(t, sessions[0].Token, session2.Token)
		assert.Equal(t, sessions[0].UserAgentDetails, "Safari / iPhone / iOS")

		assert.Equal(t, sessions[1].ID, session.ID)
		assert.Equal(t, sessions[1].Token, session.Token)
		assert.Equal(t, sessions[1].UserAgentDetails, "Safari / macOS")
	})
}
