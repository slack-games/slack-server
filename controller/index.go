package controller

import (
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/slack-games/slack-server/server"
)

var indexTemplate *template.Template

type IndexController struct {
	Context server.Context
}

func (i *IndexController) handleLoginSuccess(w http.ResponseWriter, r *http.Request) {
	if err := indexTemplate.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Register creates a new subrouter for the hangman and adds the http handlers
func (i *IndexController) Register(router *mux.Router) *mux.Router {

	indexTemplate = template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/index.html",
	))

	router.HandleFunc("/", i.handleLoginSuccess)

	return router
}
