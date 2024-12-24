package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type RedeemRequest struct {
	UserID int `json:"user_id"`
	Points int `json:"points"`
}

type RedeemResponse struct {
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	RemainingPoints int    `json:"remaining_points"`
}

//(w http.ResponseWriter, r *http.Request, db *sql.DB) {

func RedeemPointsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {

	var req RedeemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Get the user's total valid points
	var totalPoints int
	err := db.QueryRow(`
			SELECT SUM(points) FROM points
			WHERE user_id = ? AND valid_until > NOW()
		`, req.UserID).Scan(&totalPoints)
	if err != nil {
		http.Error(w, "Failed to fetch user points", http.StatusInternalServerError)
		return
	}

	// Check if user has enough points
	if req.Points > totalPoints {
		json.NewEncoder(w).Encode(RedeemResponse{
			Success: false,
			Message: "Insufficient points",
		})
		return
	}

	// Deduct points
	tx, err := db.Begin()
	if err != nil {
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
		tx.Rollback()
		http.Error(w, "Failed to fetch points for redemption", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() && pointsToRedeem > 0 {
		var id, availablePoints int
		if err := rows.Scan(&id, &availablePoints); err != nil {
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
			tx.Rollback()
			http.Error(w, "Failed to update points", http.StatusInternalServerError)
			return
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Get the remaining points
	var remainingPoints int
	err = db.QueryRow(`
			SELECT SUM(points) FROM points
			WHERE user_id = ? AND valid_until > NOW()
		`, req.UserID).Scan(&remainingPoints)
	if err != nil {
		http.Error(w, "Failed to fetch remaining points", http.StatusInternalServerError)
		return
	}

	// Respond with success
	json.NewEncoder(w).Encode(RedeemResponse{
		Success:         true,
		Message:         "Points redeemed successfully",
		RemainingPoints: remainingPoints,
	})

}

type PointsHistoryRequest struct {
	UserID          int    `json:"user_id"`
	StartDate       string `json:"start_date"`       // Format: YYYY-MM-DD
	EndDate         string `json:"end_date"`         // Format: YYYY-MM-DD
	TransactionType string `json:"transaction_type"` // Earned, Redeemed, Expired
}

type PointsHistoryResponse struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	Points          int       `json:"points"`
	TransactionType string    `json:"transaction_type"`
	TransactionDate time.Time `json:"transaction_date"`
	Reason          string    `json:"reason,omitempty"` // Optional: only for expired transactions
}

func PointsHistoryHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var req PointsHistoryRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

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
	}

	// Add filter for transaction type
	if req.TransactionType != "" {
		query += " AND transaction_type = ?"
		args = append(args, req.TransactionType)
	}

	// Execute query
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Println("Failed to fetch points history:", err)
		http.Error(w, "Failed to fetch points history", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Parse results
	var history []PointsHistoryResponse
	for rows.Next() {
		var record PointsHistoryResponse
		if err := rows.Scan(
			&record.ID, &record.UserID, &record.Points, &record.TransactionType,
			&record.TransactionDate, &record.Reason,
		); err != nil {
			log.Println("Failed to scan points history row:", err)
			http.Error(w, "Failed to process points history", http.StatusInternalServerError)
			return
		}
		history = append(history, record)
	}

	// Respond with the history
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
