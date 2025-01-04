-- +goose Up
-- +goose StatementBegin

-- Создаем перечисления для статусов пользователей и сделок
CREATE TYPE user_role AS ENUM ('ADMIN', 'USER');
CREATE TYPE user_status AS ENUM ('ACTIVE', 'BLOCKED', 'DELETED');
CREATE TYPE trade_status AS ENUM ('PENDING', 'ACCEPTED', 'DECLINED', 'CANCELLED', 'COMPLETED');

-- Таблица пользователей с расширенными полями безопасности и аудита
CREATE TABLE users (
                       user_id BIGSERIAL PRIMARY KEY,
                       username VARCHAR(255) UNIQUE NOT NULL,
                       password VARCHAR(255) NOT NULL,
                       email VARCHAR(255) UNIQUE NOT NULL,
                       role user_role NOT NULL DEFAULT 'USER',
                       status user_status NOT NULL DEFAULT 'ACTIVE',
                       balance BIGINT NOT NULL DEFAULT 0,
                       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Таблица предметов с расширенными метаданными
CREATE TABLE items (
                       item_id BIGSERIAL PRIMARY KEY,
                       item_name VARCHAR(255) NOT NULL,
                       item_price BIGINT NOT NULL,
                       rarity VARCHAR(50) NOT NULL,
                       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Таблица инвентаря для отслеживания предметов пользователей
CREATE TABLE inventory (
                           user_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
                           item_id BIGINT REFERENCES items(item_id) ON DELETE CASCADE,
                           item_count BIGINT NOT NULL DEFAULT 1,
                           PRIMARY KEY (user_id, item_id)
);

-- Таблица сделок с полным отслеживанием состояний
CREATE TABLE trades (
                        trade_id BIGSERIAL PRIMARY KEY,
                        trade_description TEXT,
                        trade_creator_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
                        trade_receiver_id BIGINT REFERENCES users(user_id) ON DELETE CASCADE,
                        trade_status trade_status NOT NULL DEFAULT 'PENDING',
                        total_price BIGINT NOT NULL,
                        trade_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Таблица предметов в сделках
CREATE TABLE trade_items (
                             trade_id BIGINT REFERENCES trades(trade_id) ON DELETE CASCADE,
                             item_id BIGINT REFERENCES items(item_id) ON DELETE CASCADE,
                             item_count BIGINT NOT NULL,
                             PRIMARY KEY (trade_id, item_id)
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_items_name ON items(item_name);
CREATE INDEX idx_items_rarity ON items(rarity);
CREATE INDEX idx_trades_status ON trades(trade_status);
CREATE INDEX idx_trades_creator ON trades(trade_creator_id);
CREATE INDEX idx_trades_receiver ON trades(trade_receiver_id);
CREATE INDEX idx_trades_date ON trades(trade_date);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Удаляем индексы
DROP INDEX IF EXISTS idx_trades_date;
DROP INDEX IF EXISTS idx_trades_receiver;
DROP INDEX IF EXISTS idx_trades_creator;
DROP INDEX IF EXISTS idx_trades_status;
DROP INDEX IF EXISTS idx_items_rarity;
DROP INDEX IF EXISTS idx_items_name;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;

-- Удаляем таблицы
DROP TABLE IF EXISTS trade_items;
DROP TABLE IF EXISTS trades;
DROP TABLE IF EXISTS inventory;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS users;

-- Удаляем типы
DROP TYPE IF EXISTS trade_status;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;

-- +goose StatementEnd
