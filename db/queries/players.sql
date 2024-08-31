-- name: GetPlayer :one
SELECT * FROM Players
WHERE PlayerID = ?;

-- name: GetPlayerByName :one
SELECT * FROM Players
WHERE Username = ?;

-- name: ListPlayers :many
SELECT * FROM Players
ORDER BY PlayerID;