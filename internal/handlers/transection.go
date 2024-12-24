package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type AddTransactionRequest struct {
	TransactionID     string  `json:"transaction_id"`
	UserID            int     `json:"user_id"`
	TransactionAmount float64 `json:"transaction_amount"`
	Category          string  `json:"category"`
	TransactionDate   string  `json:"transaction_date"`
	ProductCode       string  `json:"product_code"`
}

type AddTransactionResponse struct {
	Message string `json:"message"`
	Points  int    `json:"points"`
}

// AddTransactionHandler records a user's purchase transaction and updates loyalty points
func AddTransactionHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Parse the request body
	var req AddTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate category and calculate points multiplier
	categoryMultipliers := map[string]float64{
		"electronics": 1.0,
		"groceries":   2.0,
		"clothing":    1.5,
	}

	multiplier, ok := categoryMultipliers[req.Category]
	if !ok {
		http.Error(w, "Invalid category", http.StatusBadRequest)
		return
	}

	// Calculate points
	points := int(req.TransactionAmount * multiplier)

	// Record the transaction in the database
	query := `INSERT INTO transactions (transaction_id, user_id, transaction_amount, category, transaction_date, product_code, points) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, req.TransactionID, req.UserID, req.TransactionAmount, req.Category, req.TransactionDate, req.ProductCode, points)
	if err != nil {
		http.Error(w, "Could not record transaction", http.StatusInternalServerError)
		return
	}

	// Update the user's loyalty points balance
	_, err = db.Exec("UPDATE users SET loyalty_points = loyalty_points + ? WHERE id = ?", points, req.UserID)
	if err != nil {
		http.Error(w, "Could not update loyalty points", http.StatusInternalServerError)
		return
	}

	// Respond with success message and points
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AddTransactionResponse{
		Message: "Transaction recorded successfully",
		Points:  points,
	})
}
