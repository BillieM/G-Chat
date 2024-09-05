-- +goose Up
-- +goose StatementBegin
ALTER TABLE Players ADD AvatarExists BOOLEAN;
ALTER TABLE Players ADD AvatarLastRequested DATETIME;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE Players DROP AvatarExists;
ALTER TABLE Players DROP AvatarLastRequested;
-- +goose StatementEnd
