CREATE TABLE IF NOT EXISTS users (
    login VARCHAR(50) NOT NULL PRIMARY KEY,
    password BYTEA NOT NULL,
    current REAL,
    withdrawn REAL 
);

CREATE TABLE IF NOT EXISTS orders (
    number VARCHAR(50) NOT NULL PRIMARY KEY,
    uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    user_login VARCHAR(50) NOT NULL,
    status TEXT CHECK (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED')) DEFAULT 'NEW',
    accrual REAL,
    FOREIGN KEY (user_login) REFERENCES users(login)
);

CREATE TABLE IF NOT EXISTS withdrawals (
    number VARCHAR(50) NOT NULL PRIMARY KEY,
    processed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    user_login VARCHAR(50) NOT NULL,
    sum REAL,
    FOREIGN KEY (user_login) REFERENCES users(login)
);