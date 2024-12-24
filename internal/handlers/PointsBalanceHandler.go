package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"strconv"
)

type PointsHistory struct {
	TransactionDate string `json:"transaction_date"`
	Points          int    `json:"points"`
	Reason          string `json:"reason"`
}

type PointsBalanceResponse struct {
	Balance int             `json:"balance"`
	History []PointsHistory `json:"history"`
}

// PointsBalanceHandler returns the user's current points balance and history
func PointsBalanceHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Parse query parameters for pagination
	userID := r.URL.Query().Get("user_id")
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// Get the user's current points balance
	var balance int
	err = db.QueryRow("SELECT loyalty_points FROM users WHERE id = ?", userID).Scan(&balance)
	if err != nil {
		http.Error(w, "Could not retrieve points balance", http.StatusInternalServerError)
		return
	}

	// Get the user's points history
	history := []PointsHistory{}
	query := `SELECT transaction_date, points, category FROM transactions WHERE user_id = ? ORDER BY transaction_date DESC LIMIT ? OFFSET ?`
	rows, err := db.Query(query, userID, pageSize, offset)
	if err != nil {
		http.Error(w, "Could not retrieve points history", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var record PointsHistory
		var category string
		if err := rows.Scan(&record.TransactionDate, &record.Points, &category); err != nil {
			http.Error(w, "Error scanning points history", http.StatusInternalServerError)
			return
		}
		record.Reason = "Transaction - " + category
		history = append(history, record)
	}

	// Respond with balance and history
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PointsBalanceResponse{
		Balance: balance,
		History: history,
	})
}
