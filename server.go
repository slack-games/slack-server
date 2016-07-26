package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"image/png"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os"
	// "os"
	"regexp"
	"strconv"
	"text/template"
	"time"

	"gopkg.in/go-playground/validator.v8"

	"golang.org/x/oauth2"
	slackoauth "golang.org/x/oauth2/slack"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	"github.com/justinas/alice"
	_ "github.com/lib/pq"
	"github.com/slack-games/slack-client"
	hngcmd "github.com/slack-games/slack-hangman/commands"
	"github.com/slack-games/slack-server/datastore"
	tttcmd "github.com/slack-games/slack-tictactoe/commands"
)

var index, failTemplate, successTemplate *template.Template
var oauthConf *oauth2.Config
var config Config
var oauthState string
var decoder *schema.Decoder
var validate *validator.Validate

// Config needed config to run the application
type Config struct {
	DBUrl      string
	Port       string
	FontPath   string
	SlackToken string
	ClientID   string
	SecretKey  string
	BasePath   string
}

// CommandInput user input for the game commands
type CommandInput struct {
	ChannelName string `schema:"channel_name" validate:"required"`
	ChannelID   string `schema:"channel_id" validate:"required,alphanum"`
	TeamID      string `schema:"team_id" validate:"required,alphanum"`
	UserID      string `schema:"user_id" validate:"required,alphanum"`
	Text        string `schema:"text" validate:"required"`
	Domain      string `schema:"team_domain" validate:"required"`
	Name        string `schema:"user_name" validate:"required"`
}

func randomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
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

func (c *AppContext) isGameTTTCommandHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		command := r.PostFormValue("command")

		if command != "/ttt" {
			sendResponse(w, slack.ResponseMessage{
				Text:        "Make sure you have command set to /ttt",
				Attachments: []slack.Attachment{},
			})
			return
		}

		log.Println("Valid ttt game command found")
		next.ServeHTTP(w, r)
	})
}

func (c *AppContext) isGameHngCommandHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		command := r.PostFormValue("command")

		if command != "/hng" {
			sendResponse(w, slack.ResponseMessage{
				Text:        "Make sure you have command set to /hng",
				Attachments: []slack.Attachment{},
			})
			return
		}

		log.Println("Valid hng game command found")
		next.ServeHTTP(w, r)
	})
}

func (c *AppContext) debugFormValues(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Show keys for debugging
		log.Println("> Form debug starts here")
		for key, value := range r.PostForm {
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

		// -1 the move number as we use th indexing from 0 to 8 in development
		message = tttcmd.MoveCommand(c.db, userID, uint8(moveTo)-1)
	}

	sendResponse(w, message)
}

func (c *AppContext) hangmanGameHandler(w http.ResponseWriter, r *http.Request) {
	var message slack.ResponseMessage

	inputError := slack.ResponseMessage{
		Text: "Could not parse the game input",
	}

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		sendResponse(w, inputError)
		return
	}

	input := &CommandInput{}
	decoder.IgnoreUnknownKeys(true)

	// r.PostForm is a map of our POST form values
	err = decoder.Decode(input, r.PostForm)
	if err != nil {
		log.Println(err)
		sendResponse(w, inputError)
		return
	}

	// Validation
	err = validate.Struct(input)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		log.Println(validationErrors)
		sendResponse(w, inputError)
		return
	}

	guessRegexp, _ := regexp.Compile("^guess ([a-z])$")

	// TODO: Move the user get and create to middleware ?
	user, err := datastore.GetOrSaveNew(c.db, input.UserID, input.TeamID, input.Name, input.Domain)
	if err != nil {
		log.Fatalln("Could not save or get the user", input.UserID, err)
	}
	fmt.Println("User", user)

	switch input.Text {
	case "start":
		// Starts the new game
		message = hngcmd.StartCommand(c.db, input.UserID)
	case "current":
		// Return the current game state, with information of previous move
		message = hngcmd.CurrentCommand(c.db, input.UserID)
	case "ping":
		// Starts the new game
		message = hngcmd.PingCommand()
	}

	// Make turn on board and get back the response
	if guessRegexp.MatchString(input.Text) {
		// Second element hold character
		guess := guessRegexp.FindStringSubmatch(input.Text)[1]

		message = hngcmd.GuessCommand(c.db, input.UserID, rune(guess[0]))
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

func (c *AppContext) handleLoginSuccess(w http.ResponseWriter, r *http.Request) {
	if err := successTemplate.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *AppContext) handleLoginFail(w http.ResponseWriter, r *http.Request) {
	if err := failTemplate.Execute(w, nil); err != nil {
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

	_, err = datastore.GetTeam(c.db, response.Team.TeamID)
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

			_, err = datastore.NewTeam(c.db, team)

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

func (c *AppContext) actionHandler(w http.ResponseWriter, r *http.Request) {
	result, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Panicln("Could not dump request")
	}
	fmt.Printf("%s\n\n", result)

	log.Println("Valid slack token")
	sendResponse(w, slack.ResponseMessage{
		Text:        "Make sure the Slack APP tokens are same",
		Attachments: []slack.Attachment{},
	})
}

// Router is wrap the routes
func Router(context AppContext) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", context.handleIndex)

	r.HandleFunc("/login/success", context.handleLoginSuccess)
	r.HandleFunc("/login/fail", context.handleLoginFail)

	// Login handlers
	r.HandleFunc("/login/slack", context.handleSlackLogin)
	r.HandleFunc("/login/slack/callback", context.handleSlackCallback)

	// Game command handling path
	gameMiddleware := alice.New(
		context.debugFormValues,
		context.slackTokenHandler,
		context.isGameTTTCommandHandler,
	)
	r.Handle("/game/tictactoe", gameMiddleware.ThenFunc(context.tictactoeGameHandler))
	// Id example 95cccffc-de50-4cd8-9ac7-74b52c6f306e
	r.HandleFunc("/game/tictactoe/image/{id:\\w{8}-\\w{4}-\\w{4}-\\w{4}-\\w{12}}",
		context.tictactoeImageHandler)

	hangmanGameMiddleware := alice.New(
		context.debugFormValues,
		context.slackTokenHandler,
		context.isGameHngCommandHandler,
	)
	r.Handle("/game/hangman", hangmanGameMiddleware.ThenFunc(context.hangmanGameHandler))
	r.HandleFunc("/game/hangman/image/{id:\\w{8}-\\w{4}-\\w{4}-\\w{4}-\\w{12}}",
		context.hangmanImageHandler)

	actionsMiddleware := alice.New(context.debugFormValues)
	r.Handle("/game/interactive-messages", actionsMiddleware.ThenFunc(context.actionHandler))

	static := http.StripPrefix("/asset/", http.FileServer(http.Dir("./resource/")))
	r.PathPrefix("/asset/").Handler(static)

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
		BasePath:   os.Getenv("BASE_PATH"),
	}

	oauthConf = &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.SecretKey,
		Scopes:       []string{"team:read", "commands"},
		Endpoint:     slackoauth.Endpoint,
	}

	// random string for oauth2 API calls to protect against CSRF
	oauthState = randomString(33)

	decoder = schema.NewDecoder()

	validate = validator.New(&validator.Config{TagName: "validate"})

	index = template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/index.html",
	))

	successTemplate = template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/success.html",
	))
	failTemplate = template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/fail.html",
	))
}

func main() {
	db := sqlx.MustConnect("postgres", config.DBUrl)

	context := AppContext{db, config}
	router := Router(context)

	recoveryRouter := handlers.RecoveryHandler()(router)
	loggedRouter := handlers.LoggingHandler(os.Stdout, recoveryRouter)
	compressRouter := handlers.CompressHandler(loggedRouter)

	log.Printf("Starting server on port %s\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, compressRouter))
}
