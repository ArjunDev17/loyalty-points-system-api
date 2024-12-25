## Loyalty Points System API
A high-performance API for managing loyalty points in an e-commerce platform. Features include secure user authentication, transaction-based points accrual, redemption management, points expiration handling, and detailed points activity logs. Built with Golang and MySQL for scalability, reliability, and performance.
## Overview
The Loyalty Points System API is a backend service that manages loyalty points for users. Features include:

- User authentication (login, refresh tokens)
- Earning, redeeming, and expiring points
- Viewing points balance and transaction history
- Automated expiration of points through a scheduled background job

This application is built using **Golang** and **MySQL**.

---

## Prerequisites

Before running the application, ensure the following:

1. **Go Installed**:
   - Version: `1.19` or higher
   - Install from [Go Downloads](https://go.dev/dl/).

2. **MySQL Database**:
   - Ensure MySQL is installed and running.
   - Create the database schema as outlined below.

3. **Dependencies**:
   - Installed using `go mod tidy` (see steps below).

---

## Setting Up the Database

### Create the Database
Run the following SQL script to set up the database and tables:

```sql
CREATE DATABASE loyalty_db;
USE loyalty_db;

CREATE TABLE users (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    refresh_token VARCHAR(255),
    refresh_token_expires_at TIMESTAMP,
    loyalty_points INT DEFAULT 0
);

CREATE TABLE audit_log (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    action VARCHAR(255) NOT NULL,
    details TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE expired_points_log (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    expired_points INT NOT NULL,
    expired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE points (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    transaction_id VARCHAR(255) NOT NULL,
    points INT NOT NULL,
    transaction_type ENUM('Earned', 'Redeemed', 'Expired') NOT NULL,
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    valid_until TIMESTAMP DEFAULT NULL,
    reason VARCHAR(255) DEFAULT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE transactions (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    transaction_id VARCHAR(255) NOT NULL UNIQUE,
    user_id INT NOT NULL,
    transaction_amount DECIMAL(10, 2) NOT NULL,
    category VARCHAR(50) NOT NULL,
    transaction_date TIMESTAMP NOT NULL,
    product_code VARCHAR(255),
    points INT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### Add Test Data
Insert sample data for testing:

```sql
INSERT INTO users (username, password_hash) VALUES
('testuser', '<hashed_password>');

INSERT INTO points (user_id, points, transaction_type, transaction_date, valid_until)
VALUES
(1, 100, 'Earned', NOW(), DATE_ADD(NOW(), INTERVAL 1 YEAR));
```

Replace `<hashed_password>` with a password hash generated using bcrypt.

---

## Running the Application

### Clone the Repository
```bash
git clone https://github.com/ArjunDev17/loyalty-points-system-api
cd loyalty-points-system-api
```

### Install Dependencies
```bash
go mod tidy
```

### Set Up Environment Variables
Create environment variable files under `config/env/`.

#### Example: `config/env/dev.env`
```env
APP_PORT=8080
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=loyalty_db
JWT_SECRET=your_secret_key
POINTS_EXPIRATION_DAYS=365
```

Update the values to match your MySQL credentials and other configuration.

### Run the Application
Start the server:
```bash
go run cmd/main.go
```

### Verify the Application
1. **Health Check**:
   ```bash
   curl -X GET http://localhost:8080/health
   ```
   Response:
   ```json
   {"status": "UP"}
   ```

2. **Login**:
   ```bash
   curl -X POST http://localhost:8080/login \
   -H "Content-Type: application/json" \
   -d '{"username": "testuser", "password": "password123"}'
   ```

3. **Redeem Points**:
   ```bash
   curl -X POST http://localhost:8080/redeem \
   -H "Content-Type: application/json" \
   -d '{"user_id": 1, "points": 50}'
   ```

4. **Points History**:
   ```bash
   curl -X GET "http://localhost:8080/points-history?user_id=1&start_date=2023-01-01&end_date=2023-12-31&transaction_type=Earned"
   ```

---

## Scheduled Task: Points Expiration

The application automatically marks points as expired daily using a scheduled background job.

### Verify Expiration
1. Ensure the cron job runs as part of the application startup.
2. Check the `points` and `expired_points_log` tables for expired entries.

---

## Additional Notes

- **JWT Secret**: Use a strong, random secret for `JWT_SECRET`.
- **Database Connection**: Ensure your database credentials are correct in the environment files.
- **Testing**: Use tools like Postman, curl, or any HTTP client to test the API endpoints.

---

For any issues or contributions, feel free to raise a PR or issue on the repository!
