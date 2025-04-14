package models

import (
	"database/sql"
)

type Repos struct {
	Users    UserRepoInterface
	Sessions SessionRepoInterface
}

func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Users:    &UserRepo{db: db},
		Sessions: &SessionRepo{db: db},
	}
}
