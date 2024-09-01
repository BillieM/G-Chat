-- name: GetPlayer :one
SELECT * FROM Players
WHERE PlayerID = ?;

-- name: GetPlayerByName :one
SELECT * FROM Players
WHERE Username = ?;

-- name: ListPlayers :many
SELECT * FROM Players
ORDER BY PlayerID;

-- name: CreatePlayer :one
INSERT INTO Players (
  Username
) VALUES (
  ?
)
RETURNING *;    

-- name: UpdatePlayerUserData :one
UPDATE Players
SET FigureString = ?,
    Motto = ?,
    Membersince = ?,
    UserDataExists = 1,
    UserDataLastRequested = DATETIME('now')
WHERE PlayerID = ?
RETURNING *;

-- name: UpdatePlayerFigure :one
UPDATE Players
SET FigureExists = 1,
    FigureLastRequested = DATETIME('now') 
WHERE PlayerID = ?
RETURNING *;