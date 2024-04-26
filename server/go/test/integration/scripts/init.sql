CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- initialize users table
CREATE TABLE IF NOT EXISTS users (
    id UUID DEFAULT uuid_generate_v4(),
    username VARCHAR(30) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    points INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (id)
);

-- users view
CREATE VIEW users_view AS
SELECT id, username, points
FROM users;