CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS balances (
    id      UUID    PRIMARY KEY,
    user_id BIGINT  NOT NULL,
    value   BIGINT  NOT NULL DEFAULT 0
    CONSTRAINT uni_balances_user_id UNIQUE (user_id)
);

INSERT INTO balances (id, user_id, value)
VALUES
    (uuid_generate_v4(), 1, 1000),
    (uuid_generate_v4(), 2, 1000),
    (uuid_generate_v4(), 3, 1000)
ON CONFLICT (user_id) DO NOTHING;