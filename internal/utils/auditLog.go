package utils

import (
	"database/sql"
	"log"
)

func LogAction(db *sql.DB, userID int, action, details string) {
	go func() {
		query := `INSERT INTO audit_log (user_id, action, details, created_at) VALUES (?, ?, ?, NOW())`
		_, err := db.Exec(query, userID, action, details)
		if err != nil {
			log.Printf("Error logging audit action for user %d: %v", userID, err)
		}
	}()
}
