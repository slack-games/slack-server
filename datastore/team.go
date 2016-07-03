package datastore

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type Team struct {
	TeamID      string    `db:"team_id"`
	Name        string    `db:"name"`
	Domain      string    `db:"domain"`
	EmailDomain string    `db:"email_domain"`
	Created     time.Time `db:"created_at"`
	Modified    time.Time `db:"modified_at"`
}

func GetTeam(db *sqlx.DB, ID string) (Team, error) {
	team := Team{}

	sql := `
		SELECT *
		FROM gms.teams
		WHERE team_id = $1
		LIMIT 1
	`

	err := db.Get(&team, sql, ID)
	return team, err
}

func NewTeam(db *sqlx.DB, team Team) (sql.Result, error) {
	sql := `
		INSERT INTO gms.teams
			(team_id, name, domain, email_domain)
		VALUES
			(:team_id, :name, :domain, :email_domain)
	`
	return db.NamedExec(sql, team)
}
