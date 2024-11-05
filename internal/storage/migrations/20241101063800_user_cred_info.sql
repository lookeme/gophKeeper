-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
CREATE TABLE UsersCredInfo (
   id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),   -- Unique identifier for each
   user_id       UUID  NOT NULL,                               -- Внешний ключ для пользователя, к которому привязана карта
   name          VARCHAR(50) NOT NULL,                         -- Имя информации
   data          VARCHAR(256),                                 -- Зашифрованная Инфо
   type          INT NOT NULL,                                 -- Тип хранимой инфы (cart, creds)
   version       INT GENERATED ALWAYS AS IDENTITY NOT NULL,    -- Версия
   created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,          -- Дата создания учетной записи
   updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,          -- Дата обновления учетной записи
   CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES Users(id)

);
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
