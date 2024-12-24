package utils

import (
	"database/sql"
	"log"
	"time"
)

// LogAction represents an action to be logged
func LogAction(db *sql.DB, userID int, action, details string) {
	_, err := db.Exec(
		`INSERT INTO audit_log (user_id, action, details, timestamp) VALUES (?, ?, ?, ?)`,
		userID, action, details, time.Now(),
	)
	if err != nil {
		log.Println("Failed to log action to audit log:", err)
	}
}
