
# Slack Games

Contains the HTTP glue code for the multiple games also the OAuth login to register
commands and team.


## Installation

Dev database deploy command:

```
psql -h localhost -p 5432 -U user -P password -f data/deploy.sql
```

Example of the Heroku DB init:
```
heroku pg:psql --app brain-games < data/deploy.sql
```


## Config

Env variables to change:

```
# Server port
PORT=8080
# Font path for the image drawings
FONT_PATH=./resource/font
# Image path for the hangman
IMAGE_PATH=./resource/images

# Postgres database connection url
DB_URL=postgres://dsfdsfdsfds:@localhost:5432/postgres?sslmode=disable
# For testing Postgres instance
DB_TEST=postgres://dsfdsfdsfdsfds:@localhost:5432/postgres?sslmode=disable

# Slack tokens
CLIENT_ID=21321321321.21321321321
SECRET_KEY=dsfdsfds76afc938f54399231321321

# Verification token from Slack registration
APP_TOKEN=dsfaferwafergdfsrtgh
```


# TODO

  - [x] Add the move numbers to TTT board
  - Validation for the POST input
  - Fix the redirect urls and base path for different envs


## Queries
```
SELECT
    S.*, U.name AS first, U2.name AS second
FROM
    gms.states AS S
INNER JOIN
    gms.users AS U ON S.first_user_id = U.user_id
INNER JOIN
    gms.users as U2 ON S.second_user_id = U2.user_id;
``
