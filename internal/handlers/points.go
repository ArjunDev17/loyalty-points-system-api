package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"loyalty-points-system-api/internal/models"
	"net/http"
)

// (w http.ResponseWriter, r *http.Request, db *sql.DB) {
func RedeemPointsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("RedeemPointsHandler: Starting to process redeem points request.")

	var req models.RedeemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("RedeemPointsHandler: Received request to redeem %d points for user %d.", req.Points, req.UserID)

	// Get the user's total valid points
	var totalPoints int
	err := db.QueryRow(`
			SELECT COALESCE(SUM(points), 0) FROM points
			WHERE user_id = ? AND valid_until > NOW()
		`, req.UserID).Scan(&totalPoints)
	if err != nil {
		log.Printf("Error fetching user points for user %d: %v", req.UserID, err)
		http.Error(w, "Failed to fetch user points", http.StatusInternalServerError)
		return
	}

	log.Printf("RedeemPointsHandler: User %d has %d total points.", req.UserID, totalPoints)

	// Check if user has enough points
	if req.Points > totalPoints {
		log.Printf("Insufficient points for user %d: requested %d, available %d.", req.UserID, req.Points, totalPoints)
		json.NewEncoder(w).Encode(models.RedeemResponse{
			Success: false,
			Message: "Insufficient points",
		})
		return
	}

	// Deduct points
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting database transaction: %v", err)
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}

	pointsToRedeem := req.Points
	rows, err := tx.Query(`
			SELECT id, points FROM points
			WHERE user_id = ? AND valid_until > NOW()
			ORDER BY valid_until ASC
		`, req.UserID)
	if err != nil {
		log.Printf("Error fetching points for redemption for user %d: %v", req.UserID, err)
		tx.Rollback()
		http.Error(w, "Failed to fetch points for redemption", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() && pointsToRedeem > 0 {
		var id, availablePoints int
		if err := rows.Scan(&id, &availablePoints); err != nil {
			log.Printf("Error scanning points row for user %d: %v", req.UserID, err)
			tx.Rollback()
			http.Error(w, "Error processing points", http.StatusInternalServerError)
			return
		}

		if availablePoints <= pointsToRedeem {
			// Use all points from this row
			_, err = tx.Exec(`DELETE FROM points WHERE id = ?`, id)
			pointsToRedeem -= availablePoints
		} else {
			// Use partial points and update the row
			_, err = tx.Exec(`UPDATE points SET points = points - ? WHERE id = ?`, pointsToRedeem, id)
			pointsToRedeem = 0
		}

		if err != nil {
			log.Printf("Error updating points for user %d in row %d: %v", req.UserID, id, err)
			tx.Rollback()
			http.Error(w, "Failed to update points", http.StatusInternalServerError)
			return
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction for user %d: %v", req.UserID, err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Get the remaining points
	var remainingPoints int
	err = db.QueryRow(`
			SELECT COALESCE(SUM(points), 0) FROM points
			WHERE user_id = ? AND valid_until > NOW()
		`, req.UserID).Scan(&remainingPoints)
	if err != nil {
		log.Printf("Error fetching remaining points for user %d: %v", req.UserID, err)
		http.Error(w, "Failed to fetch remaining points", http.StatusInternalServerError)
		return
	}

	log.Printf("RedeemPointsHandler: Successfully redeemed points for user %d. Remaining points: %d.", req.UserID, remainingPoints)

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.RedeemResponse{
		Success:         true,
		Message:         "Points redeemed successfully",
		RemainingPoints: remainingPoints,
	})
}

func PointsHistoryHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("PointsHistoryHandler: Starting to process points history request.")

	var req models.PointsHistoryRequest
	if r.Body == nil {
		log.Println("PointsHistoryHandler: Request body is empty.")
		http.Error(w, "Request body cannot be empty", http.StatusBadRequest)
		return
	}

	// Decode request body
	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("PointsHistoryHandler: Received request for user_id: %d, start_date: %s, end_date: %s, transaction_type: %s",
		req.UserID, req.StartDate, req.EndDate, req.TransactionType)

	// Base query
	query := `
		SELECT id, user_id, points, transaction_type, transaction_date, reason
		FROM points
		WHERE user_id = ?
	`
	args := []interface{}{req.UserID}

	// Add filters for date range
	if req.StartDate != "" && req.EndDate != "" {
		query += " AND transaction_date BETWEEN ? AND ?"
		args = append(args, req.StartDate, req.EndDate)
		log.Printf("PointsHistoryHandler: Filtering points history between %s and %s", req.StartDate, req.EndDate)
	}

	// Add filter for transaction type
	if req.TransactionType != "" {
		query += " AND transaction_type = ?"
		args = append(args, req.TransactionType)
		log.Printf("PointsHistoryHandler: Filtering points history by transaction_type: %s", req.TransactionType)
	}

	// Execute query
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error executing query to fetch points history: %v", err)
		http.Error(w, "Failed to fetch points history", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Parse results
	var history []models.PointsHistoryResponse
	for rows.Next() {
		var record models.PointsHistoryResponse
		if err := rows.Scan(
			&record.ID, &record.UserID, &record.Points, &record.TransactionType,
			&record.TransactionDate, &record.Reason,
		); err != nil {
			log.Printf("Error scanning points history row: %v", err)
			http.Error(w, "Failed to process points history", http.StatusInternalServerError)
			return
		}
		history = append(history, record)
	}

	// Check for errors during row iteration
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v", err)
		http.Error(w, "Failed to process points history", http.StatusInternalServerError)
		return
	}

	// Respond with the history
	log.Printf("PointsHistoryHandler: Successfully retrieved %d records for user_id: %d", len(history), req.UserID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
