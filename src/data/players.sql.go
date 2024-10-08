// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: players.sql

package data

import (
	"context"
	"database/sql"
	"strings"
)

const createPlayer = `-- name: CreatePlayer :one
INSERT INTO Players (
  Username
) VALUES (
  ?
)
RETURNING playerid, username, figurestring, motto, membersince, userdataexists, figureexists, userdatalastrequested, figurelastrequested, AvatarExists, AvatarLastRequested, IsMe
`

func (q *Queries) CreatePlayer(ctx context.Context, username string) (Player, error) {
	row := q.db.QueryRowContext(ctx, createPlayer, username)
	var i Player
	err := row.Scan(
		&i.Playerid,
		&i.Username,
		&i.Figurestring,
		&i.Motto,
		&i.Membersince,
		&i.Userdataexists,
		&i.Figureexists,
		&i.Userdatalastrequested,
		&i.Figurelastrequested,
		&i.AvatarExists,
		&i.AvatarLastRequested,
		&i.IsMe,
	)
	return i, err
}

const getPlayer = `-- name: GetPlayer :one
SELECT playerid, username, figurestring, motto, membersince, userdataexists, figureexists, userdatalastrequested, figurelastrequested, AvatarExists, AvatarLastRequested, IsMe FROM Players
WHERE PlayerID = ?
`

func (q *Queries) GetPlayer(ctx context.Context, playerid int64) (Player, error) {
	row := q.db.QueryRowContext(ctx, getPlayer, playerid)
	var i Player
	err := row.Scan(
		&i.Playerid,
		&i.Username,
		&i.Figurestring,
		&i.Motto,
		&i.Membersince,
		&i.Userdataexists,
		&i.Figureexists,
		&i.Userdatalastrequested,
		&i.Figurelastrequested,
		&i.AvatarExists,
		&i.AvatarLastRequested,
		&i.IsMe,
	)
	return i, err
}

const getPlayerByName = `-- name: GetPlayerByName :one
SELECT playerid, username, figurestring, motto, membersince, userdataexists, figureexists, userdatalastrequested, figurelastrequested, AvatarExists, AvatarLastRequested, IsMe FROM Players
WHERE Username = ?
`

func (q *Queries) GetPlayerByName(ctx context.Context, username string) (Player, error) {
	row := q.db.QueryRowContext(ctx, getPlayerByName, username)
	var i Player
	err := row.Scan(
		&i.Playerid,
		&i.Username,
		&i.Figurestring,
		&i.Motto,
		&i.Membersince,
		&i.Userdataexists,
		&i.Figureexists,
		&i.Userdatalastrequested,
		&i.Figurelastrequested,
		&i.AvatarExists,
		&i.AvatarLastRequested,
		&i.IsMe,
	)
	return i, err
}

const listPlayers = `-- name: ListPlayers :many
SELECT playerid, username, figurestring, motto, membersince, userdataexists, figureexists, userdatalastrequested, figurelastrequested, AvatarExists, AvatarLastRequested, IsMe FROM Players
ORDER BY PlayerID
`

func (q *Queries) ListPlayers(ctx context.Context) ([]Player, error) {
	rows, err := q.db.QueryContext(ctx, listPlayers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Player
	for rows.Next() {
		var i Player
		if err := rows.Scan(
			&i.Playerid,
			&i.Username,
			&i.Figurestring,
			&i.Motto,
			&i.Membersince,
			&i.Userdataexists,
			&i.Figureexists,
			&i.Userdatalastrequested,
			&i.Figurelastrequested,
			&i.AvatarExists,
			&i.AvatarLastRequested,
			&i.IsMe,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listPlayersByUsernames = `-- name: ListPlayersByUsernames :many
SELECT playerid, username, figurestring, motto, membersince, userdataexists, figureexists, userdatalastrequested, figurelastrequested, AvatarExists, AvatarLastRequested, IsMe FROM Players
WHERE Username IN (/*SLICE:usernames*/?)
`

func (q *Queries) ListPlayersByUsernames(ctx context.Context, usernames []string) ([]Player, error) {
	query := listPlayersByUsernames
	var queryParams []interface{}
	if len(usernames) > 0 {
		for _, v := range usernames {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:usernames*/?", strings.Repeat(",?", len(usernames))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:usernames*/?", "NULL", 1)
	}
	rows, err := q.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Player
	for rows.Next() {
		var i Player
		if err := rows.Scan(
			&i.Playerid,
			&i.Username,
			&i.Figurestring,
			&i.Motto,
			&i.Membersince,
			&i.Userdataexists,
			&i.Figureexists,
			&i.Userdatalastrequested,
			&i.Figurelastrequested,
			&i.AvatarExists,
			&i.AvatarLastRequested,
			&i.IsMe,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updatePlayerAvatar = `-- name: UpdatePlayerAvatar :one
UPDATE Players
SET AvatarExists = 1,
    AvatarLastRequested = DATETIME('now') 
WHERE PlayerID = ?
RETURNING playerid, username, figurestring, motto, membersince, userdataexists, figureexists, userdatalastrequested, figurelastrequested, AvatarExists, AvatarLastRequested, IsMe
`

func (q *Queries) UpdatePlayerAvatar(ctx context.Context, playerid int64) (Player, error) {
	row := q.db.QueryRowContext(ctx, updatePlayerAvatar, playerid)
	var i Player
	err := row.Scan(
		&i.Playerid,
		&i.Username,
		&i.Figurestring,
		&i.Motto,
		&i.Membersince,
		&i.Userdataexists,
		&i.Figureexists,
		&i.Userdatalastrequested,
		&i.Figurelastrequested,
		&i.AvatarExists,
		&i.AvatarLastRequested,
		&i.IsMe,
	)
	return i, err
}

const updatePlayerFigure = `-- name: UpdatePlayerFigure :one
UPDATE Players
SET FigureExists = 1,
    FigureLastRequested = DATETIME('now') 
WHERE PlayerID = ?
RETURNING playerid, username, figurestring, motto, membersince, userdataexists, figureexists, userdatalastrequested, figurelastrequested, AvatarExists, AvatarLastRequested, IsMe
`

func (q *Queries) UpdatePlayerFigure(ctx context.Context, playerid int64) (Player, error) {
	row := q.db.QueryRowContext(ctx, updatePlayerFigure, playerid)
	var i Player
	err := row.Scan(
		&i.Playerid,
		&i.Username,
		&i.Figurestring,
		&i.Motto,
		&i.Membersince,
		&i.Userdataexists,
		&i.Figureexists,
		&i.Userdatalastrequested,
		&i.Figurelastrequested,
		&i.AvatarExists,
		&i.AvatarLastRequested,
		&i.IsMe,
	)
	return i, err
}

const updatePlayerFigureString = `-- name: UpdatePlayerFigureString :one
UPDATE Players
SET FigureString = ?
WHERE PlayerID = ?
RETURNING playerid, username, figurestring, motto, membersince, userdataexists, figureexists, userdatalastrequested, figurelastrequested, AvatarExists, AvatarLastRequested, IsMe
`

type UpdatePlayerFigureStringParams struct {
	Figurestring sql.NullString
	Playerid     int64
}

func (q *Queries) UpdatePlayerFigureString(ctx context.Context, arg UpdatePlayerFigureStringParams) (Player, error) {
	row := q.db.QueryRowContext(ctx, updatePlayerFigureString, arg.Figurestring, arg.Playerid)
	var i Player
	err := row.Scan(
		&i.Playerid,
		&i.Username,
		&i.Figurestring,
		&i.Motto,
		&i.Membersince,
		&i.Userdataexists,
		&i.Figureexists,
		&i.Userdatalastrequested,
		&i.Figurelastrequested,
		&i.AvatarExists,
		&i.AvatarLastRequested,
		&i.IsMe,
	)
	return i, err
}

const updatePlayerSetIsMe = `-- name: UpdatePlayerSetIsMe :one
UPDATE Players 
SET IsMe = 1
WHERE PlayerID = ?
RETURNING playerid, username, figurestring, motto, membersince, userdataexists, figureexists, userdatalastrequested, figurelastrequested, AvatarExists, AvatarLastRequested, IsMe
`

func (q *Queries) UpdatePlayerSetIsMe(ctx context.Context, playerid int64) (Player, error) {
	row := q.db.QueryRowContext(ctx, updatePlayerSetIsMe, playerid)
	var i Player
	err := row.Scan(
		&i.Playerid,
		&i.Username,
		&i.Figurestring,
		&i.Motto,
		&i.Membersince,
		&i.Userdataexists,
		&i.Figureexists,
		&i.Userdatalastrequested,
		&i.Figurelastrequested,
		&i.AvatarExists,
		&i.AvatarLastRequested,
		&i.IsMe,
	)
	return i, err
}

const updatePlayerSetNotMe = `-- name: UpdatePlayerSetNotMe :one
UPDATE Players 
SET IsMe = 1
WHERE PlayerID = ?
RETURNING playerid, username, figurestring, motto, membersince, userdataexists, figureexists, userdatalastrequested, figurelastrequested, AvatarExists, AvatarLastRequested, IsMe
`

func (q *Queries) UpdatePlayerSetNotMe(ctx context.Context, playerid int64) (Player, error) {
	row := q.db.QueryRowContext(ctx, updatePlayerSetNotMe, playerid)
	var i Player
	err := row.Scan(
		&i.Playerid,
		&i.Username,
		&i.Figurestring,
		&i.Motto,
		&i.Membersince,
		&i.Userdataexists,
		&i.Figureexists,
		&i.Userdatalastrequested,
		&i.Figurelastrequested,
		&i.AvatarExists,
		&i.AvatarLastRequested,
		&i.IsMe,
	)
	return i, err
}

const updatePlayerUserData = `-- name: UpdatePlayerUserData :one
UPDATE Players
SET Motto = ?,
    Membersince = ?,
    UserDataExists = 1,
    UserDataLastRequested = DATETIME('now')
WHERE PlayerID = ?
RETURNING playerid, username, figurestring, motto, membersince, userdataexists, figureexists, userdatalastrequested, figurelastrequested, AvatarExists, AvatarLastRequested, IsMe
`

type UpdatePlayerUserDataParams struct {
	Motto       sql.NullString
	Membersince sql.NullTime
	Playerid    int64
}

func (q *Queries) UpdatePlayerUserData(ctx context.Context, arg UpdatePlayerUserDataParams) (Player, error) {
	row := q.db.QueryRowContext(ctx, updatePlayerUserData, arg.Motto, arg.Membersince, arg.Playerid)
	var i Player
	err := row.Scan(
		&i.Playerid,
		&i.Username,
		&i.Figurestring,
		&i.Motto,
		&i.Membersince,
		&i.Userdataexists,
		&i.Figureexists,
		&i.Userdatalastrequested,
		&i.Figurelastrequested,
		&i.AvatarExists,
		&i.AvatarLastRequested,
		&i.IsMe,
	)
	return i, err
}
