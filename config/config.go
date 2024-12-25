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

func LoadConfig(env string) *Config {
	err := godotenv.Load("config/env/" + env + ".env")
	if err != nil {
		log.Fatalf("Error loading %s environment file: %v", env, err)
	}

	expirationDays, _ := strconv.Atoi(os.Getenv("POINTS_EXPIRATION_DAYS"))

	return &Config{
		AppPort:              os.Getenv("APP_PORT"),
		DBHost:               os.Getenv("DB_HOST"),
		DBPort:               os.Getenv("DB_PORT"),
		DBUser:               os.Getenv("DB_USER"),
		DBPassword:           os.Getenv("DB_PASSWORD"),
		DBName:               os.Getenv("DB_NAME"),
		JWTSecret:            os.Getenv("JWT_SECRET"),
		PointsExpirationDays: expirationDays,
	}
}

func ConnectDB(cfg *Config) *sql.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=30s&readTimeout=30s&writeTimeout=30s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	db.SetMaxOpenConns(10) // Reduced for better stability
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(1 * time.Minute)
	db.SetConnMaxIdleTime(30 * time.Second)

	// Verify connection
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
