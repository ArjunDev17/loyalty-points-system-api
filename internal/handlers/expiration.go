package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"loyalty-points-system-api/internal/utils"
)

// ExpirePoints runs periodically to mark points as expired
func ExpirePoints(db *sql.DB) {
	log.Println("Starting points expiration job...")

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Println("Failed to start transaction:", err)
		return
	}

	// Find points that have expired
	rows, err := tx.Query(`
		SELECT id, user_id, points FROM points
		WHERE valid_until < NOW() AND transaction_type = 'Earned'
	`)
	if err != nil {
		log.Println("Failed to query expired points:", err)
		tx.Rollback()
		return
	}
	defer rows.Close()

	// Process each expired point entry
	for rows.Next() {
		var id, userID, expiredPoints int
		if err := rows.Scan(&id, &userID, &expiredPoints); err != nil {
			log.Println("Failed to scan expired points row:", err)
			tx.Rollback()
			return
		}

		// Log the expired points in the points table
		_, err := tx.Exec(`
			UPDATE points SET transaction_type = 'Expired', reason = 'Expired'
			WHERE id = ?
		`, id)
		if err != nil {
			log.Println("Failed to mark points as expired:", err)
			tx.Rollback()
			return
		}

		// Log the expired points in a separate table
		_, err = tx.Exec(`
			INSERT INTO expired_points_log (user_id, expired_points)
			VALUES (?, ?)
		`, userID, expiredPoints)
		if err != nil {
			log.Println("Failed to log expired points:", err)
			tx.Rollback()
			return
		}

		// Deduct expired points from user's balance (optional if using sum queries)
		log.Printf("User %d: Expired %d points", userID, expiredPoints)
		// Log expired points in audit log
		utils.LogAction(db, userID, "Expire Points", fmt.Sprintf("Expired %d points", expiredPoints))
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Println("Failed to commit transaction:", err)
		return
	}

	log.Println("Points expiration job completed successfully.")
}
