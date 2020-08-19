CREATE TABLE IF NOT EXISTS customers (
    id UUID NOT NULL PRIMARY KEY,
    first_name VARCHAR(256),
    last_name VARCHAR(256),
    email VARCHAR(256),
    phone VARCHAR(256)
);