package models

import (
	"database/sql"
)

type Repos struct {
	Users    UserRepoInterface
	Sessions SessionRepoInterface
	Sites    SiteRepoInterface
	Events   EventRepoInterface
}

func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Users:    &UserRepo{db: db},
		Sessions: &SessionRepo{db: db},
		Sites:    &SiteRepo{db: db},
		Events:   &EventRepo{db: db},
	}
}
