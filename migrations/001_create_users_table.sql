CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add column for refresh token
ALTER TABLE users ADD COLUMN refresh_token VARCHAR(255);

-- Add column for refresh token expiration time
ALTER TABLE users ADD COLUMN refresh_token_expires_at TIMESTAMP;


CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    refresh_token VARCHAR(255), -- To store the refresh token
    refresh_token_expires_at TIMESTAMP, -- To store the expiration time of the refresh token
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
