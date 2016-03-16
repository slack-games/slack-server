package main

import (
	"encoding/json"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"text/template"
	"time"

	"golang.org/x/oauth2"
	slackoauth "golang.org/x/oauth2/slack"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	"github.com/justinas/alice"
	_ "github.com/lib/pq"
	"github.com/riston/slack-client"
	"github.com/riston/slack-server/datastore"
	"github.com/riston/slack-tictactoe/commands"
	tttdatastore "github.com/riston/slack-tictactoe/datastore"
	drawBoard "github.com/riston/slack-tictactoe/draw"
)

var index *template.Template
var oauthConf *oauth2.Config
var config Config
var oauthState string

// Config needed config to run the application
type Config struct {
	DBUrl      string
	Port       string
	SlackToken string
	ClientID   string
	SecretKey  string
}

// AppContext holds reference example for database instance
type AppContext struct {
	db     *sqlx.DB
	config Config
}

func (c *AppContext) slackTokenHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		slackToken := r.PostFormValue("token")

		if c.config.SlackToken != slackToken {
			sendResponse(w, slack.ResponseMessage{
				Text:        "Make sure the Slack APP tokens are same",
				Attachments: []slack.Attachment{},
			})
			return
		}
		log.Println("Valid slack token")
		next.ServeHTTP(w, r)
	})
}

func (c *AppContext) isGameCommandHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		command := r.PostFormValue("command")

		if command != "/game" {
			sendResponse(w, slack.ResponseMessage{
				Text:        "Make sure you have command set to /game",
				Attachments: []slack.Attachment{},
			})
			return
		}

		log.Println("Valid game command found")
		next.ServeHTTP(w, r)
	})
}

func (c *AppContext) gameHandler(w http.ResponseWriter, r *http.Request) {
	var message slack.ResponseMessage

	// Authentication, check the token, team id and user id
	text := r.PostFormValue("text")

	domain := r.PostFormValue("team_domain")
	teamID := r.PostFormValue("team_id")
	userID := r.PostFormValue("user_id")
	name := r.PostFormValue("user_name")

	log.Println("> Headers start here")
	// Show keys for debugging
	for key, value := range r.Form {
		log.Printf("\tkey: %s value: %s\n", key, value)
	}
	log.Println("> Headers ends here")

	moveRegexp, _ := regexp.Compile("^move (\\d)$")

	user, err := datastore.GetOrSaveNew(c.db, userID, teamID, name, domain)
	if err != nil {
		log.Fatalln("Could not save or get the user", userID, err)
	}

	fmt.Println("Res", user)

	switch text {
	case "start":
		// Starts the new game
		message = commands.StartCommand(c.db, userID)

	case "current":
		// Return the current game state, with information of previous move
		// and also with current whose turn it is
		message = commands.CurrentCommand(c.db, userID)

	case "stats":
		// Get the players stats
		// Not implemented yet
	case "help":
		// Trigger also the command help
		message = commands.HelpCommand()

	case "ping":
		// Test if the commands are responding
		message = commands.PingCommand()

	default:
		// Show the help message, also list all possible commands available
		// Did not found your command, did you mean this ?
		message = commands.HelpCommand()
	}

	// Make turn on board and get back the response
	if moveRegexp.MatchString(text) {

		// Second element hold number
		strNumber := moveRegexp.FindStringSubmatch(text)[1]
		moveTo, _ := strconv.ParseInt(strNumber, 10, 8)

		message = commands.MoveCommand(c.db, userID, uint8(moveTo))
	}

	sendResponse(w, message)
}

func sendResponse(w http.ResponseWriter, message slack.ResponseMessage) {
	// Set headers
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	// Generate message output
	if err := json.NewEncoder(w).Encode(message); err != nil {
		log.Fatal("Could not generate the message json", err)
	}
}

func (c *AppContext) stateImageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	state, err := tttdatastore.GetState(c.db, id)
	if err != nil {
		http.Error(w, "Could not get the state", 404)
		return
	}

	ttt := tttdatastore.CreateTicTacToeBoard(state)

	img := drawBoard.Draw(ttt)
	err = png.Encode(w, img)
	if err != nil {
		http.Error(w, "Could not save the image", 500)
		return
	}
}

func (c *AppContext) handleIndex(w http.ResponseWriter, r *http.Request) {
	if err := index.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *AppContext) handleSlackLogin(w http.ResponseWriter, r *http.Request) {
	url := oauthConf.AuthCodeURL(oauthState, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (c *AppContext) handleSlackCallback(w http.ResponseWriter, r *http.Request) {
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

	_, err = datastore.NewTeam(c.db, datastore.Team{
		TeamID:      response.Team.TeamID,
		Name:        response.Team.Name,
		Domain:      response.Team.Domain,
		EmailDomain: response.Team.EmailDomain,
		Created:     time.Now(),
		Modified:    time.Now(),
	})

	if err != nil {
		log.Printf("Failed to save the team result %v\n", err)
		http.Redirect(w, r, "/login/fail", http.StatusTemporaryRedirect)
		return
	}

	log.Println("Team saved ", response.Team)

	http.Redirect(w, r, "/login/success", http.StatusTemporaryRedirect)
}

// Router is wrap the routes
func Router(context AppContext) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", context.handleIndex)

	// Login handlers
	r.HandleFunc("/login/slack", context.handleSlackLogin)
	r.HandleFunc("/login/slack/callback", context.handleSlackCallback)

	// Game command handling path
	gameMiddleware := alice.New(context.slackTokenHandler, context.isGameCommandHandler)
	r.Handle("/game", gameMiddleware.ThenFunc(context.gameHandler))

	// Id example 95cccffc-de50-4cd8-9ac7-74b52c6f306e
	r.HandleFunc("/image/{id:\\w{8}-\\w{4}-\\w{4}-\\w{4}-\\w{12}}", context.stateImageHandler)
	return r
}

func init() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalln("Make sure the PORT variable has been set")
	}

	DBUrl := os.Getenv("DB_URL")
	if DBUrl == "" {
		log.Fatalln("No database URL provided, could not continue")
	}

	config = Config{
		DBUrl:      DBUrl,
		Port:       port,
		SlackToken: os.Getenv("APP_TOKEN"),
		ClientID:   os.Getenv("CLIENT_ID"),
		SecretKey:  os.Getenv("SECRET_KEY"),
	}

	oauthConf = &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.SecretKey,
		Scopes:       []string{"team:read", "commands"},
		Endpoint:     slackoauth.Endpoint,
	}
	// random string for oauth2 API calls to protect against CSRF
	oauthState = "thisshouldberandom"

	index = template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/index.html",
	))
}

func main() {
	db := sqlx.MustConnect("postgres", config.DBUrl)

	context := AppContext{db, config}
	router := Router(context)

	recoveryRouter := handlers.RecoveryHandler()(router)
	loggedRouter := handlers.LoggingHandler(os.Stdout, recoveryRouter)
	compressRouter := handlers.CompressHandler(loggedRouter)

	log.Println("Starting server ...")
	log.Fatal(http.ListenAndServe(":"+config.Port, compressRouter))
}
