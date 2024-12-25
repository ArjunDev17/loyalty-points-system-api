CREATE TABLE transactions (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    transaction_id VARCHAR(255) NOT NULL UNIQUE,
    user_id INT NOT NULL,
    transaction_amount DECIMAL(10, 2) NOT NULL,
    category VARCHAR(50) NOT NULL,
    transaction_date TIMESTAMP NOT NULL,
    product_code VARCHAR(255),
    points INT NOT NULL
);
