package views

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/justinas/nosurf"
	"github.com/robyparr/event-horizon/internal"
	"github.com/robyparr/event-horizon/internal/models"
)

//go:embed "html" "static"
var Files embed.FS

type viewModel struct {
	IsAuthenticated bool
	CurrentUser     *models.User
	Form            any
	Flash           map[string]string
	Data            map[string]any
	CSRFToken       string
	AppStartedAt    int64
}

func NewViewModel(app *internal.App, r *http.Request, form any) viewModel {
	user, _ := app.GetCurrentUser(r)

	return viewModel{
		IsAuthenticated: app.IsAuthenticated(r),
		CurrentUser:     &user,
		Flash:           app.GetFlashDataFromContext(r),
		Data:            make(map[string]any),
		AppStartedAt:    app.StartedAt.Unix(),

		CSRFToken: nosurf.Token(r),
		Form:      form,
	}
}

func CompileViews() (map[string]*template.Template, error) {
	templates := map[string]*template.Template{}

	pages, err := getTemplatePaths(Files, "html/pages", ".tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name, _ := strings.CutPrefix(page, "html/pages/")

		patterns := []string{
			"html/base.html.tmpl",
			"html/partials/*.tmpl",
			page,
		}

		ts, err := template.New(name).Funcs(functions).ParseFS(Files, patterns...)
		if err != nil {
			return nil, err
		}

		templates[name] = ts
	}

	return templates, nil
}

func getTemplatePaths(fileSys fs.FS, rootDir string, ext string) ([]string, error) {
	var paths []string
	err := fs.WalkDir(fileSys, rootDir, func(path string, d fs.DirEntry, err error) error {
		if filepath.Ext(path) == ext {
			paths = append(paths, path)
		}

		return nil
	})

	return paths, err
}
