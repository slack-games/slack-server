package datastore

import (
	"database/sql"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

type User struct {
	UserID     string    `db:"user_id"`
	TeamID     string    `db:"team_id"`
	Name       string    `db:"name"`
	TeamDomain string    `db:"team_domain"`
	Created    time.Time `db:"created_at"`
	Modified   time.Time `db:"modified_at"`
}

func GetUser(db *sqlx.DB, ID string) (User, error) {
	user := User{}

	sql := `
		SELECT *
		FROM gms.users
		WHERE user_id = $1
		LIMIT 1
	`

	err := db.Get(&user, sql, ID)
	return user, err
}

func NewUser(db *sqlx.DB, user User) (sql.Result, error) {
	sql := `
		INSERT INTO gms.users
			(user_id, team_id, name, team_domain)
		VALUES
			(:user_id, :team_id, :name, :team_domain)
	`
	return db.NamedExec(sql, user)
}

func GetAll(db *sqlx.DB) ([]User, error) {
	users := []User{}
	sql := `SELECT * FROM gms.users`
	err := db.Select(&users, sql)

	return users, err
}

func GetOrSaveNew(db *sqlx.DB, userID, teamID, name, domain string) (User, error) {
	user, err := GetUser(db, userID)
	if err != nil {
		// No rows try to create a new user
		if err == sql.ErrNoRows {
			user := User{
				userID,
				teamID,
				name,
				domain,
				time.Now(),
				time.Now(),
			}

			log.Println("Create a new user", user)
			result, err := NewUser(db, user)
			if err != nil {
				log.Fatalln("Could not create a new user", err)
				return User{}, err
			}

			if rows, _ := result.RowsAffected(); rows != 1 {
				log.Fatalln("Failed to create a new user")
				return User{}, err
			}
		} else {
			log.Fatalln("Could not get the user from DB", userID)
			return User{}, err
		}
	}
	return user, nil
}
