package handlers

import (
	"fmt"
	"net/http"

	"github.com/robyparr/event-horizon/internal"
	"github.com/robyparr/event-horizon/internal/models"
	"github.com/robyparr/event-horizon/internal/validator"
	"github.com/robyparr/event-horizon/internal/views"
)

type newSiteForm struct {
	Name string `form:"name"`
	validator.Validator
}

func sitesListHandler(app *internal.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.MustGetCurrentUser(r)
		sites, err := app.Repos.Sites.ListForUser(&user)
		if err != nil {
			app.ServerError(w, r, err)
			return
		}

		vm := views.NewViewModel(app, r, newSiteForm{})
		vm.Data["sites"] = sites
		app.Render(w, r, http.StatusOK, "sites/index.html.tmpl", vm)
	})
}

func sitesCreateHandler(app *internal.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var form newSiteForm
		err := app.DecodePostForm(r, &form)
		if err != nil {
			app.ServerError(w, r, err)
			return
		}

		form.CheckField(form.Name != "", "name", "Name is required")
		if !form.Valid() {
			vm := views.NewViewModel(app, r, form)
			app.Render(w, r, http.StatusUnprocessableEntity, "sites/index.html.tmpl", vm)
			return
		}

		site := models.Site{Name: form.Name, UserID: app.MustGetCurrentUser(r).ID}
		err = app.Repos.Sites.Insert(&site)
		if err != nil {
			app.ServerError(w, r, err)
			return
		}

		app.SetFlash(w, "info", "Site created successfully.")
		http.Redirect(w, r, fmt.Sprintf("/sites/%d", site.ID), http.StatusSeeOther)
	})
}

func sitesShowHandler(app *internal.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := readIDParam(r)
		if err != nil {
			app.ServerError(w, r, err)
			return
		}

		user := app.MustGetCurrentUser(r)
		site, err := app.Repos.Sites.FindForUser(&user, id)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		vm := views.NewViewModel(app, r, nil)
		vm.Data["site"] = site
		app.Render(w, r, http.StatusOK, "sites/show.html.tmpl", vm)
	})
}

func sitesDeleteHandler(app *internal.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := readIDParam(r)
		if err != nil {
			app.ServerError(w, r, err)
			return
		}

		user := app.MustGetCurrentUser(r)
		site, err := app.Repos.Sites.FindForUser(&user, id)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		err = app.Repos.Sites.Delete(&site)
		if err != nil {
			app.ServerError(w, r, err)
			return
		}

		app.SetFlash(w, "info", fmt.Sprintf("'%s' has been deleted.", site.Name))
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}
