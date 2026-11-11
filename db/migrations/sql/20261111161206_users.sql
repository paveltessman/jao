-- +goose Up
CREATE TABLE app_user (
    id text PRIMARY KEY NOT NULL,
    name text UNIQUE NOT NULL,
    lastseen timestamp NOT NULL,
    active boolean NOT NULL,
    passhash text NOT NULL,
    refresh_token_id text NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
);

-- +goose Down
DROP TABLE app_user;
