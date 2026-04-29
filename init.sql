CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS balances (
    id      UUID    PRIMARY KEY,
    user_id BIGINT  NOT NULL UNIQUE,
    value   BIGINT  NOT NULL DEFAULT 0
);

INSERT INTO balances (id, user_id, value)
VALUES
    (uuid_generate_v4(), 1, 1000),
    (uuid_generate_v4(), 2, 1000),
    (uuid_generate_v4(), 3, 1000)
ON CONFLICT (user_id) DO NOTHING;