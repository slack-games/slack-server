package controller

import (
	"fmt"
	"image/png"
	"log"
	"net/http"
	"regexp"

	"gopkg.in/go-playground/validator.v8"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/justinas/alice"
	"github.com/slack-games/slack-client"
	hngcmd "github.com/slack-games/slack-hangman/commands"
	"github.com/slack-games/slack-server/datastore"
	"github.com/slack-games/slack-server/server"
)

var decoder = schema.NewDecoder()

// HangmanController hangman controller
type HangmanController struct {
	Context server.Context
}

func (h *HangmanController) isGameCommandHandler(next http.Handler) http.Handler {
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

func (h *HangmanController) hangmanGameHandler(w http.ResponseWriter, r *http.Request) {
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
	err = h.Context.Validate.Struct(input)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		log.Println(validationErrors)
		sendResponse(w, inputError)
		return
	}

	guessRegexp, _ := regexp.Compile("^guess ([a-z])$")

	// TODO: Move the user get and create to middleware ?
	user, err := datastore.GetOrSaveNew(h.Context.Db, input.UserID, input.TeamID, input.Name, input.Domain)
	if err != nil {
		log.Fatalln("Could not save or get the user", input.UserID, err)
	}
	fmt.Println("User", user)

	switch input.Text {
	case "start":
		// Starts the new game
		message = hngcmd.StartCommand(h.Context.Db, input.UserID)
	case "current":
		// Return the current game state, with information of previous move
		message = hngcmd.CurrentCommand(h.Context.Db, input.UserID)
	case "ping":
		// Starts the new game
		message = hngcmd.PingCommand()
	}

	// Make turn on board and get back the response
	if guessRegexp.MatchString(input.Text) {
		// Second element hold character
		guess := guessRegexp.FindStringSubmatch(input.Text)[1]

		message = hngcmd.GuessCommand(h.Context.Db, input.UserID, rune(guess[0]))
	}

	sendResponse(w, message)
}

func (h *HangmanController) getImageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	image, err := hngcmd.GetGameImage(h.Context.Db, id)
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

// Register creates a new subrouter for the hangman and adds the http handlers
func (h *HangmanController) Register(router *mux.Router) *mux.Router {
	decoder.IgnoreUnknownKeys(true)
	tttRouter := router.PathPrefix("/hangman").Subrouter()

	tttRouter.HandleFunc("/image/{id:\\w{8}-\\w{4}-\\w{4}-\\w{4}-\\w{12}}", h.getImageHandler).
		Methods("GET")

	hangmanGameMiddleware := alice.New(
		slackTokenHandler(h.Context.Config.SlackToken),
		debugFormValues,
		h.isGameCommandHandler,
	)

	tttRouter.Methods("POST").
		Handler(hangmanGameMiddleware.ThenFunc(h.hangmanGameHandler))

	return tttRouter
}
