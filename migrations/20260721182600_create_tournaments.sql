-- +goose Up
SELECT
    'up SQL query';

CREATE TABLE IF NOT EXISTS (
    name text NOT NULL,
    discipline text NOT NULL,
    bracket_format text NOT NULL,
    max_pariticipants int NOT NULL,
    start_date timestamp NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
) 

-- +goose Down
SELECT
    'down SQL query';