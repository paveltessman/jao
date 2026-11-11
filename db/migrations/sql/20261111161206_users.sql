-- +goose Up
CREATE TABLE app_user (
    id text PRIMARY KEY NOT NULL,
    name text UNIQUE NOT NULL,
    lastseen integer NOT NULL,
    active boolean NOT NULL,
    passhash text NOT NULL,
    refresh_token_id text NOT NULL,
    created_at integer NOT NULL,
    updated_at integer NOT NULL
);

-- +goose Down
DROP TABLE app_user;
