package mocks

import (
	"fmt"
	"time"

	"github.com/robyparr/event-horizon/internal/models"
)

type SiteRepo struct {
	Sites map[int64]models.Site
}

func NewSiteRepo() *SiteRepo {
	return &SiteRepo{Sites: make(map[int64]models.Site)}
}

func (r *SiteRepo) ListForUser(user *models.User) ([]models.Site, error) {
	var out []models.Site
	for _, s := range r.Sites {
		if s.UserID == user.ID {
			out = append(out, s)
		}
	}

	return out, nil
}

func (r *SiteRepo) Insert(s *models.Site) error {
	if s.UserID == 0 {
		return fmt.Errorf("Invalid UserID of 0")
	}

	s.ID = int64(len(r.Sites) + 1)
	s.CreatedAt = time.Now().UTC()
	s.UpdatedAt = s.CreatedAt

	r.Sites[s.ID] = *s
	return nil
}

func (r *SiteRepo) FindForUser(u *models.User, id int64) (models.Site, error) {
	for _, s := range r.Sites {
		if s.ID == id {
			return s, nil
		}
	}

	return models.Site{}, models.ErrNoRecord
}

func (r *SiteRepo) Delete(s *models.Site) error {
	_, ok := r.Sites[s.ID]
	if !ok {
		return fmt.Errorf("No site with ID %d", s.ID)
	}

	delete(r.Sites, s.ID)
	return nil
}

func (r *SiteRepo) FindByToken(token string) (models.Site, error) {
	for _, site := range r.Sites {
		if site.Token == token {
			return site, nil
		}
	}

	return models.Site{}, models.ErrNoRecord
}
