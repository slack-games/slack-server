package main

import (
	"log"
	"net/http"
	"os"

	"gopkg.in/bluesuncorp/validator.v8"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/slack-games/slack-server/controller"
	"github.com/slack-games/slack-server/server"
)

// Router is wrap the routes
func Router(context server.Context) *mux.Router {
	router := mux.NewRouter()

	// Create game subrouter
	gameRouter := router.PathPrefix("/game").Subrouter()

	hangmanController := controller.HangmanController{Context: context}
	hangmanController.Register(gameRouter)

	tictactoeController := controller.TictactoeController{Context: context}
	tictactoeController.Register(gameRouter)

	loginController := controller.LoginController{Context: context}
	loginController.Register(router)

	indexController := controller.IndexController{Context: context}
	indexController.Register(router)

	// Static assets
	static := http.StripPrefix("/asset/", http.FileServer(http.Dir("./resource/")))
	router.PathPrefix("/asset/").Handler(static)

	return router
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalln("Make sure the PORT variable has been set")
	}

	DBUrl := os.Getenv("DB_URL")
	if DBUrl == "" {
		log.Fatalln("No database URL provided, could not continue")
	}

	config := server.Config{
		DBUrl:      DBUrl,
		Port:       port,
		FontPath:   os.Getenv("FONT_PATH"),
		SlackToken: os.Getenv("APP_TOKEN"),
		ClientID:   os.Getenv("CLIENT_ID"),
		SecretKey:  os.Getenv("SECRET_KEY"),
		BasePath:   os.Getenv("BASE_PATH"),
	}

	validate := validator.New(&validator.Config{TagName: "validate"})

	db := sqlx.MustConnect("postgres", config.DBUrl)

	context := server.Context{
		Db:       db,
		Validate: validate,
		Config:   config,
	}
	router := Router(context)

	recoveryRouter := handlers.RecoveryHandler()(router)
	loggedRouter := handlers.LoggingHandler(os.Stdout, recoveryRouter)
	compressRouter := handlers.CompressHandler(loggedRouter)

	log.Printf("Starting server on port %s\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, compressRouter))
}
