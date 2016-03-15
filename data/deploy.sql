
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Games schema, common tables between games
CREATE SCHEMA IF NOT EXISTS gms;
-- TicTacToe game schema
CREATE SCHEMA IF NOT EXISTS ttt;
-- Hangman game schema
CREATE SCHEMA IF NOT EXISTS hng;

-- DROP TABLE IF EXISTS gms.teams;
CREATE TABLE IF NOT EXISTS gms.teams (
    team_id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    domain TEXT NOT NULL,
    email_domain TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    modified_at TIMESTAMP NOT NULL DEFAULT now()
);

-- DROP TABLE IF EXISTS gms.users;
CREATE TABLE IF NOT EXISTS gms.users (
    user_id TEXT PRIMARY KEY,
    team_id TEXT,
    name TEXT NOT NULL,
    team_domain TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    modified_at TIMESTAMP NOT NULL DEFAULT now()
);

-- NB! Make sure to remove this
DROP TYPE IF EXISTS ttt.mode CASCADE;
CREATE TYPE ttt.mode AS ENUM ('Start', 'Win', 'Draw', 'GameOver', 'Turn', 'Unkown');

DROP TABLE IF EXISTS ttt.states;
CREATE TABLE IF NOT EXISTS ttt.states (
    state_id UUID PRIMARY KEY UNIQUE DEFAULT gen_random_uuid(),
    state TEXT,
    turn TEXT REFERENCES gms.users (user_id),
    mode gms.mode,
    first_user_id TEXT REFERENCES gms.users (user_id),
    second_user_id TEXT REFERENCES gms.users (user_id),
    parent_state_id UUID DEFAULT '00000000-0000-0000-0000-000000000000',
    created_at TIMESTAMP NOT NULL DEFAULT now()
);
