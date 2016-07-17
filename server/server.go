package server

import (
	"github.com/jmoiron/sqlx"
	"gopkg.in/bluesuncorp/validator.v8"
)

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

// Context holds reference example for database instance
type Context struct {
	Db       *sqlx.DB
	Validate *validator.Validate
	Config   Config
}
