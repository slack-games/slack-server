package controller

import (
	"fmt"
	"image/png"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/riston/slack-server/datastore"
	"github.com/slack-games/slack-client"
	"github.com/slack-games/slack-server/server"
	tttcmd "github.com/slack-games/slack-tictactoe/commands"
)

// TictactoeController controller
type TictactoeController struct {
	Context server.Context
}

func (t *TictactoeController) isGameCommandHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		command := r.PostFormValue("command")

		if command != "/ttt" {
			sendResponse(w, slack.ResponseMessage{
				Text: "Make sure you have command set to /ttt",
			})
			return
		}

		log.Println("Valid hng game command found")
		next.ServeHTTP(w, r)
	})
}

func (t *TictactoeController) tictactoeImageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	image, err := tttcmd.GetGameImage(t.Context.Db, id)
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

func (t *TictactoeController) gameHandler(w http.ResponseWriter, r *http.Request) {
	var message slack.ResponseMessage

	// Authentication, check the token, team id and user id
	text := r.PostFormValue("text")
	domain := r.PostFormValue("team_domain")
	teamID := r.PostFormValue("team_id")
	userID := r.PostFormValue("user_id")
	name := r.PostFormValue("user_name")

	moveRegexp, _ := regexp.Compile("^move (\\d)$")

	user, err := datastore.GetOrSaveNew(t.Context.Db, userID, teamID, name, domain)
	if err != nil {
		log.Fatalln("Could not save or get the user", userID, err)
	}

	fmt.Println("User", user)

	switch text {
	case "start":
		// Starts the new game
		message = tttcmd.StartCommand(t.Context.Db, userID)

	case "current":
		// Return the current game state, with information of previous move
		// and also with current whose turn it is
		message = tttcmd.CurrentCommand(t.Context.Db, userID)

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
		message = tttcmd.MoveCommand(t.Context.Db, userID, uint8(moveTo)-1)
	}

	sendResponse(w, message)
}

// Register adds the tictactoe routes
func (t *TictactoeController) Register(router *mux.Router) *mux.Router {
	tttRouter := router.PathPrefix("/tictactoe").Subrouter()

	tttRouter.HandleFunc("/image/{id:\\w{8}-\\w{4}-\\w{4}-\\w{4}-\\w{12}}", t.tictactoeImageHandler)

	gameMiddleware := alice.New(
		slackTokenHandler(t.Context.Config.SlackToken),
		debugFormValues,
		t.isGameCommandHandler,
	)

	tttRouter.Methods("POST").
		Handler(gameMiddleware.ThenFunc(t.gameHandler))

	return tttRouter
}
