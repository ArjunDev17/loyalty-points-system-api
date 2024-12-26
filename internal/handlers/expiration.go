package handlers

import (
	"database/sql"
	"log"
)

// ExpirePoints runs periodically to mark points as expired
func ExpirePoints(db *sql.DB) {
	log.Println("ExpirePoints: Starting points expiration job...")

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Println("Failed to start transaction:", err)
		return
	}
	defer tx.Rollback()

	// Debugging: Check how many rows match the condition
	var count int
	err = tx.QueryRow(`
		SELECT COUNT(*) FROM points
		WHERE valid_until < NOW() AND transaction_type = 'Earned'
	`).Scan(&count)
	if err != nil {
		log.Println("Error fetching count of expired points:", err)
		return
	}
	log.Printf("ExpirePoints: Found %d expired points to process.", count)

	// Find points that have expired
	rows, err := tx.Query(`
		SELECT id, user_id, points FROM points
		WHERE valid_until < NOW() AND transaction_type = 'Earned'
	`)
	if err != nil {
		log.Println("Failed to query expired points:", err)
		return
	}
	defer rows.Close()

	// Process each expired point entry
	for rows.Next() {
		var id, userID, expiredPoints int
		if err := rows.Scan(&id, &userID, &expiredPoints); err != nil {
			log.Println("Failed to scan expired points row:", err)
			return
		}
		log.Printf("ExpirePoints: Processing row: ID=%d, UserID=%d, Points=%d", id, userID, expiredPoints)

		// Update points table
		_, err = tx.Exec(`
			UPDATE points SET transaction_type = 'Expired', reason = 'Expired'
			WHERE id = ?
		`, id)
		if err != nil {
			log.Printf("Failed to mark points as expired for ID=%d: %v", id, err)
			return
		}

		// Deduct points from user's balance
		_, err = tx.Exec(`
			UPDATE users SET loyalty_points = loyalty_points - ?
			WHERE id = ?
		`, expiredPoints, userID)
		if err != nil {
			log.Printf("Failed to update user's loyalty points for UserID=%d: %v", userID, err)
			return
		}

		// Log the expired points
		_, err = tx.Exec(`
			INSERT INTO expired_points_log (user_id, expired_points, expired_at)
			VALUES (?, ?, NOW())
		`, userID, expiredPoints)
		if err != nil {
			log.Printf("Failed to log expired points for UserID=%d: %v", userID, err)
			return
		}

		log.Printf("ExpirePoints: Expired %d points for UserID=%d (Row ID=%d)", expiredPoints, userID, id)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Println("Failed to commit transaction:", err)
		return
	}
	log.Println("ExpirePoints: Points expiration job completed successfully.")
}
