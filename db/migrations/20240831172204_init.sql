-- +goose Up
-- +goose StatementBegin
CREATE TABLE Players (
	PlayerID INTEGER PRIMARY KEY,
	Username TEXT NOT NULL,
    FigureString TEXT,
    Motto TEXT,
    Membersince DATETIME,
    UserDataExists BOOLEAN,
    FigureExists BOOLEAN,
	UserDataLastRequested DATETIME,
	FigureLastRequested DATETIME
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE players;
-- +goose StatementEnd
