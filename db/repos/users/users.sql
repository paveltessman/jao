-- name: UserGetByID :one
SELECT
    id,
    name,
    lastseen,
    active,
    passhash,
    refresh_token_id,
    created_at,
    updated_at
FROM
    app_user
WHERE
    id = ?;

-- name: UserUpdate :exec
UPDATE
    app_user
SET
    lastseen = ?,
    active = ?,
    passhash = ?,
    refresh_token_id = ?,
    updated_at = ?
WHERE
    id = ?;

-- name: UserInsert :exec
INSERT INTO app_user (id, name, lastseen, active, passhash, refresh_token_id, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?);
