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
	hngcmd "github.com/riston/slack-hangman/commands"
	"github.com/riston/slack-server/datastore"
	tttcmd "github.com/riston/slack-tictactoe/commands"
)

var index *template.Template
var oauthConf *oauth2.Config
var config Config
var oauthState string

// Config needed config to run the application
type Config struct {
	DBUrl      string
	Port       string
	FontPath   string
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

func (c *AppContext) debugFormValues(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Show keys for debugging
		log.Println("> Form debug starts here")
		for key, value := range r.Form {
			log.Printf("\tkey: %s value: %s\n", key, value)
		}
		log.Println("> Headers ends here")
		next.ServeHTTP(w, r)
	})
}

func (c *AppContext) tictactoeGameHandler(w http.ResponseWriter, r *http.Request) {
	var message slack.ResponseMessage

	// Authentication, check the token, team id and user id
	text := r.PostFormValue("text")
	domain := r.PostFormValue("team_domain")
	teamID := r.PostFormValue("team_id")
	userID := r.PostFormValue("user_id")
	name := r.PostFormValue("user_name")

	moveRegexp, _ := regexp.Compile("^move (\\d)$")

	user, err := datastore.GetOrSaveNew(c.db, userID, teamID, name, domain)
	if err != nil {
		log.Fatalln("Could not save or get the user", userID, err)
	}

	fmt.Println("User", user)

	switch text {
	case "start":
		// Starts the new game
		message = tttcmd.StartCommand(c.db, userID)

	case "current":
		// Return the current game state, with information of previous move
		// and also with current whose turn it is
		message = tttcmd.CurrentCommand(c.db, userID)

	case "stats":
		// Get the players stats
		// Not implemented yet
	case "help":
		// Trigger also the command help
		message = tttcmd.HelpCommand()

	case "ping":
		// Test if the commands are responding
		message = tttcmd.PingCommand()

	default:
		// Show the help message, also list all possible commands available
		// Did not found your command, did you mean this ?
		message = tttcmd.HelpCommand()
	}

	// Make turn on board and get back the response
	if moveRegexp.MatchString(text) {

		// Second element hold number
		strNumber := moveRegexp.FindStringSubmatch(text)[1]
		moveTo, _ := strconv.ParseInt(strNumber, 10, 8)

		message = tttcmd.MoveCommand(c.db, userID, uint8(moveTo))
	}

	sendResponse(w, message)
}

func (c *AppContext) hangmanGameHandler(w http.ResponseWriter, r *http.Request) {
	var message slack.ResponseMessage

	// Authentication, check the token, team id and user id
	text := r.PostFormValue("text")
	domain := r.PostFormValue("team_domain")
	teamID := r.PostFormValue("team_id")
	userID := r.PostFormValue("user_id")
	name := r.PostFormValue("user_name")

	guessRegexp, _ := regexp.Compile("^guess ([a-z])$")

	// TODO: Move the user get and create to middleware ?
	user, err := datastore.GetOrSaveNew(c.db, userID, teamID, name, domain)
	if err != nil {
		log.Fatalln("Could not save or get the user", userID, err)
	}
	fmt.Println("User", user)

	switch text {
	case "start":
		// Starts the new game
		message = hngcmd.StartCommand(c.db, userID)
	case "current":
		// Return the current game state, with information of previous move
		message = hngcmd.CurrentCommand(c.db, userID)
	case "ping":
		// Starts the new game
		message = hngcmd.PingCommand()
	}

	// Make turn on board and get back the response
	if guessRegexp.MatchString(text) {
		// Second element hold character
		guess := guessRegexp.FindStringSubmatch(text)[1]

		message = hngcmd.GuessCommand(c.db, userID, rune(guess[0]))
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

func (c *AppContext) tictactoeImageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	image, err := tttcmd.GetGameImage(c.db, id)
	if err != nil {
		http.Error(w, "Could not get the state", 404)
		return
	}

	err = png.Encode(w, image)
	if err != nil {
		http.Error(w, "Could not save the image", 500)
		return
	}
}

func (c *AppContext) hangmanImageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	image, err := hngcmd.GetGameImage(c.db, id)
	if err != nil {
		http.Error(w, "Could not get the state", 404)
		return
	}

	err = png.Encode(w, image)
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
	r.Handle("/game/tictactoe", gameMiddleware.ThenFunc(context.tictactoeGameHandler))

	hangmanGameMiddleware := alice.New(context.debugFormValues)
	// Id example 95cccffc-de50-4cd8-9ac7-74b52c6f306e
	r.HandleFunc("/game/tictactoe/image/{id:\\w{8}-\\w{4}-\\w{4}-\\w{4}-\\w{12}}",
		context.tictactoeImageHandler)

	r.Handle("/game/hangman", hangmanGameMiddleware.ThenFunc(context.hangmanGameHandler))
	r.HandleFunc("/game/hangman/image/{id:\\w{8}-\\w{4}-\\w{4}-\\w{4}-\\w{12}}",
		context.hangmanImageHandler)

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
		FontPath:   os.Getenv("FONT_PATH"),
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
