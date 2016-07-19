package controller

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	slackoauth "golang.org/x/oauth2/slack"

	"github.com/gorilla/mux"
	"github.com/slack-games/slack-client"
	"github.com/slack-games/slack-server/datastore"
	"github.com/slack-games/slack-server/server"
	"golang.org/x/oauth2"
)

var oauthConf *oauth2.Config
var oauthState string
var failTemplate, successTemplate *template.Template

type LoginController struct {
	Context server.Context
}

func (l *LoginController) handleSlackLogin(w http.ResponseWriter, r *http.Request) {
	url := oauthConf.AuthCodeURL(oauthState, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (l *LoginController) handleSlackCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthState {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthState, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	log.Println("Received token", token.AccessToken)

	oauthClient := oauthConf.Client(oauth2.NoContext, token)
	response, err := slack.GetTeamInfo(oauthClient, token)
	if err != nil {
		log.Printf("Could not get the team info %v\n", err)
		http.Redirect(w, r, "/login/fail", http.StatusTemporaryRedirect)
		return
	}

	_, err = datastore.GetTeam(l.Context.Db, response.Team.TeamID)
	if err != nil {
		// No result found, save team information
		if err == sql.ErrNoRows {
			team := datastore.Team{
				TeamID:      response.Team.TeamID,
				Name:        response.Team.Name,
				Domain:      response.Team.Domain,
				EmailDomain: response.Team.EmailDomain,
				Created:     time.Now(),
				Modified:    time.Now(),
			}

			_, err = datastore.NewTeam(l.Context.Db, team)

			if err != nil {
				log.Printf("Failed to save the team result %v\n", err)
				http.Redirect(w, r, "/login/fail", http.StatusTemporaryRedirect)
				return
			}
			log.Println("New team created")

		} else {
			log.Printf("Could not find or save the team data %v\n", err)
			log.Println(err)
			http.Redirect(w, r, "/login/fail", http.StatusTemporaryRedirect)
			return
		}
	}

	log.Println("Team ", response.Team)
	http.Redirect(w, r, "/login/success", http.StatusTemporaryRedirect)
}

func (l *LoginController) handleLoginSuccess(w http.ResponseWriter, r *http.Request) {
	if err := successTemplate.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (l *LoginController) handleLoginFail(w http.ResponseWriter, r *http.Request) {
	if err := failTemplate.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Register creates a new subrouter for the hangman and adds the http handlers
func (l *LoginController) Register(router *mux.Router) *mux.Router {

	oauthState = randomString(33)

	oauthConf = &oauth2.Config{
		ClientID:     l.Context.Config.ClientID,
		ClientSecret: l.Context.Config.SecretKey,
		Scopes:       []string{"team:read", "commands"},
		Endpoint:     slackoauth.Endpoint,
	}

	successTemplate = template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/success.html",
	))

	failTemplate = template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/fail.html",
	))

	loginRouter := router.PathPrefix("/login").Subrouter()

	// Login handlers
	loginRouter.HandleFunc("/slack", l.handleSlackLogin)
	loginRouter.HandleFunc("/slack/callback", l.handleSlackCallback)

	loginRouter.HandleFunc("/success", l.handleLoginSuccess)
	loginRouter.HandleFunc("/fail", l.handleLoginFail)

	return loginRouter
}
