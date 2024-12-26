package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Config struct {
	AppPort              string
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	JWTSecret            string
	PointsExpirationDays int
}

// LoadConfig loads configuration from .env file
func LoadConfig(env string) (*Config, error) {
	err := godotenv.Load("config/env/" + env + ".env")
	if err != nil {
		return nil, fmt.Errorf("error loading %s environment file: %w", env, err)
	}

	expirationDays, err := strconv.Atoi(getEnv("POINTS_EXPIRATION_DAYS", "30")) // Default to 30 days
	if err != nil {
		return nil, fmt.Errorf("invalid POINTS_EXPIRATION_DAYS: %w", err)
	}

	return &Config{
		AppPort:              getEnv("APP_PORT", "8080"),
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "3306"),
		DBUser:               getEnv("DB_USER", "root"),
		DBPassword:           getEnv("DB_PASSWORD", ""),
		DBName:               getEnv("DB_NAME", "loyalty_db"),
		JWTSecret:            getEnv("JWT_SECRET", "defaultsecret"),
		PointsExpirationDays: expirationDays,
	}, nil
}

// ConnectDB connects to the database
func ConnectDB(cfg *Config) *sql.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=30s&readTimeout=30s&writeTimeout=30s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(1 * time.Minute)
	db.SetConnMaxIdleTime(30 * time.Second)

	// Retry database ping with exponential backoff
	for i := 0; i < 3; i++ {
		if err := db.Ping(); err != nil {
			log.Printf("Failed to ping database (attempt %d/3): %v", i+1, err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		log.Println("Connected to the MySQL database successfully.")
		return db
	}

	log.Fatalf("Could not establish stable connection to the database after 3 attempts")
	return nil
}

// Helper to get environment variables with default fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
