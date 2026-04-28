CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

INSERT INTO balances (id, user_id, value)
VALUES
    (uuid_generate_v4(), 1, 1000),
    (uuid_generate_v4(), 2, 1000),
    (uuid_generate_v4(), 3, 1000);