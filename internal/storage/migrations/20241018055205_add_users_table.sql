-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
CREATE TABLE Users (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Unique identifier for each user
   username VARCHAR(255) NOT NULL UNIQUE,         -- Username, must be unique
   email VARCHAR(255) NOT NULL UNIQUE,            -- Email, must be unique
   password VARCHAR(255) NOT NULL,                -- Hashed password for authentication
   role VARCHAR(50) DEFAULT 'user',               -- Role of the user (e.g., admin, user)
   created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- Timestamp of user creation
   updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP -- Timestamp of last update
);

CREATE TABLE Files (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),  -- Unique identifier for each file
       file_name VARCHAR(255) NOT NULL,                -- Name of the file
       file_path TEXT NOT NULL,                        -- Path or URL where the file is stored
       file_size BIGINT,                               -- Size of the file in bytes
       file_type VARCHAR(50),                          -- MIME type of the file (e.g., image/jpeg, application/pdf)
       owner_id UUID NOT NULL,                         -- Foreign key referencing the Users table (who uploaded the file)
       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- Timestamp when the file was uploaded
       updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- Timestamp when the file was last modified
       CONSTRAINT fk_owner FOREIGN KEY (owner_id) REFERENCES Users(id) ON DELETE CASCADE -- Ensures that the owner must exist
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
