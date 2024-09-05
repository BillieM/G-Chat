-- name: GetPlayer :one
SELECT * FROM Players
WHERE PlayerID = ?;

-- name: GetPlayerByName :one
SELECT * FROM Players
WHERE Username = ?;

-- name: ListPlayers :many
SELECT * FROM Players
ORDER BY PlayerID;

-- name: ListPlayersByUsernames :many
SELECT * FROM Players
WHERE Username IN (sqlc.slice('usernames'));

-- name: CreatePlayer :one
INSERT INTO Players (
  Username
) VALUES (
  ?
)
RETURNING *;    

-- name: UpdatePlayerUserData :one
UPDATE Players
SET Motto = ?,
    Membersince = ?,
    UserDataExists = 1,
    UserDataLastRequested = DATETIME('now')
WHERE PlayerID = ?
RETURNING *;

-- name: UpdatePlayerFigureString :one
UPDATE Players
SET FigureString = ?
WHERE PlayerID = ?
RETURNING *;

-- name: UpdatePlayerFigure :one
UPDATE Players
SET FigureExists = 1,
    FigureLastRequested = DATETIME('now') 
WHERE PlayerID = ?
RETURNING *;

-- name: UpdatePlayerAvatar :one
UPDATE Players
SET AvatarExists = 1,
    AvatarLastRequested = DATETIME('now') 
WHERE PlayerID = ?
RETURNING *;

-- name: UpdatePlayerSetIsMe :one
UPDATE Players 
SET IsMe = 1
WHERE PlayerID = ?
RETURNING *;

-- name: UpdatePlayerSetNotMe :one
UPDATE Players 
SET IsMe = 1
WHERE PlayerID = ?
RETURNING *;