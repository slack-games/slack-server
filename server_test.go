package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/slack-games/slack-client"
	"github.com/slack-games/slack-server/server"
)

var db *sqlx.DB
var myConfig server.Config

func SetUpDatabase(db *sqlx.DB) {
	//load the latest schema
	_, err := sqlx.LoadFile(db, "data/deploy.sql")

	if err != nil {
		log.Fatalln("Failed to load schema", err)
	}

	// load our test data
	_, err = sqlx.LoadFile(db, "data/fixtures.sql")

	if err != nil {
		TearDownDatabase(db)
		log.Fatalln("Failed to load fixtures", err)
	}
}

func TearDownDatabase(db *sqlx.DB) {
	// destroy all tables to reset sequences and remove test data
	_, err := sqlx.LoadFile(db, "data/cleanup.sql")

	if err != nil {
		log.Fatalln("failed to delete fixtures", err)
	}
}

func Request(values url.Values) (*slack.ResponseMessage, error) {
	r, err := http.NewRequest("POST", "/game", bytes.NewBufferString(values.Encode()))
	if err != nil {
		return nil, err
	}
	r.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded;",
	)
	w := httptest.NewRecorder()
	context := server.Context{
		Db:     db,
		Config: myConfig,
	}
	Router(context).ServeHTTP(w, r)

	var response *slack.ResponseMessage
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func TestMain(m *testing.M) {
	myConfig = server.Config{
		DBUrl: os.Getenv("DB_TEST"),
	}

	db = sqlx.MustConnect("postgres", os.Getenv("DB_TEST"))
	fmt.Println("Setup DB")
	SetUpDatabase(db)
	res := m.Run()

	fmt.Println("Tear down DB")
	TearDownDatabase(db)
	os.Exit(res)
}

func TestRandomPage(t *testing.T) {
	r, _ := http.NewRequest("POST", "/randompage", nil)
	r.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)
	w := httptest.NewRecorder()
	context := server.Context{
		Db:     db,
		Config: config,
	}
	Router(context).ServeHTTP(w, r)

	if w.Code != 404 {
		t.Error("Random page should return 404")
	}
}

func TestPingPage(t *testing.T) {
	data := url.Values{}
	data.Set("command", "/ping")

	response, err := Request(data)
	if err != nil {
		t.Error("Could not make a request ", err)
	}

	if response.Text != "This is a ping page" {
		t.Error("Make sure the ping page text matches")
	}
}

func TestNewUserRegistration(t *testing.T) {
	data := url.Values{}
	data.Set("command", "/game")
	data.Set("text", "move 3")
	// data.Set("user_id", "U000000000")
	data.Set("user_name", "Mike")
	data.Set("team_domain", "smarts")

	response, err := Request(data)
	if err != nil {
		t.Error("Could not make a request ", err)
	}

	fmt.Println("Response", response)
}

// func TestServerResponse(t *testing.T) {
// 	data := url.Values{}
// 	data.Set("command", "/game")
// 	data.Set("text", "move 3")
//
// 	response, err := Request(data)
// 	if err != nil {
// 		t.Error("Could not make a request ", err)
// 	}
//
// 	fmt.Println(response)
// }

func TestUserRecord(t *testing.T) {
	fmt.Println("Test user related queries")
}
