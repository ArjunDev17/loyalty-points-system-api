CREATE TABLE points (
    id INT AUTO_INCREMENT PRIMARY KEY,              -- Unique identifier for the point record
    user_id INT NOT NULL,                           -- Foreign key to associate with a user
    points INT NOT NULL,                            -- Number of points
    transaction_type ENUM('Earned', 'Redeemed', 'Expired') NOT NULL, -- Type of transaction
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Date of the transaction
    valid_until TIMESTAMP DEFAULT NULL,            -- Expiration date of the points
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE -- Ensures points are deleted if the user is deleted
);
CREATE TABLE points (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    transaction_id VARCHAR(255) NOT NULL,
    points INT NOT NULL,
    transaction_type ENUM('Earned', 'Redeemed', 'Expired') NOT NULL,
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    valid_until TIMESTAMP,
    reason VARCHAR(255)
);
