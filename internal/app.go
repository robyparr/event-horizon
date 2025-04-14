package internal

import (
	"html/template"
	"log/slog"
	"sync"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/gorilla/securecookie"
	"github.com/robyparr/event-horizon/internal/models"
)

const (
	EnvDevelopment string = "development"
	EnvProduction  string = "production"
)

type App struct {
	Logger       *slog.Logger
	Repos        *models.Repos
	Views        map[string]*template.Template
	FormDecoder  *form.Decoder
	SecureCookie *securecookie.SecureCookie
	Config       Config
	StartedAt    time.Time
	WG           sync.WaitGroup
}

type Config struct {
	Environment string
	Host        string
	DatabaseURL string
	SkipTLS     bool
}

func (c Config) IsProductionEnv() bool {
	return c.Environment == EnvProduction
}
