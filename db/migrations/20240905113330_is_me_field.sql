-- +goose Up
-- +goose StatementBegin
ALTER TABLE Players ADD IsMe BOOLEAN NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE Players DROP IsMe;
-- +goose StatementEnd
