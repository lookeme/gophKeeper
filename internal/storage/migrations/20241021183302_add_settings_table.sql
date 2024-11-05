-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
CREATE TABLE Settings (
   id         UUID PRIMARY KEY UNIQUE DEFAULT gen_random_uuid(),         -- Unique identifier for each user
   key        VARCHAR(255) NOT NULL,                                     -- key, must be unique
   value      VARCHAR(255) NOT NULL,                                     -- value
   version    INT DEFAULT 1 NOT NULL,                                    -- Версия кредов
   created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,        -- Timestamp of user creation
   updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP         -- Timestamp of last update
);
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
