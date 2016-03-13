
# Slack Games

## Installation

## Config

Env variables to change:

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
