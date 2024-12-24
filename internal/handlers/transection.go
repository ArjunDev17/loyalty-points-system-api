package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"loyalty-points-system-api/internal/models"
	utils "loyalty-points-system-api/internal/utils"
	"net/http"
)

// AddTransactionHandler records a user's purchase transaction and logs the action
func AddTransactionHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Parse the request body
	var req models.AddTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Error decoding request payload:", err)
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
		log.Println("Invalid category provided:", req.Category)
		http.Error(w, "Invalid category", http.StatusBadRequest)
		return
	}

	// Calculate points
	points := int(req.TransactionAmount * multiplier)
	log.Printf("Calculated %d points for user %d in category %s", points, req.UserID, req.Category)

	// Record the transaction in the database
	query := `INSERT INTO transactions (transaction_id, user_id, transaction_amount, category, transaction_date, product_code, points) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, req.TransactionID, req.UserID, req.TransactionAmount, req.Category, req.TransactionDate, req.ProductCode, points)
	if err != nil {
		log.Println("Error recording transaction:", err)
		http.Error(w, "Could not record transaction", http.StatusInternalServerError)
		return
	}
	log.Printf("Transaction %s recorded successfully for user %d", req.TransactionID, req.UserID)

	// Update the user's loyalty points balance
	_, err = db.Exec("UPDATE users SET loyalty_points = loyalty_points + ? WHERE id = ?", points, req.UserID)
	if err != nil {
		log.Println("Error updating loyalty points for user:", req.UserID, "Error:", err)
		http.Error(w, "Could not update loyalty points", http.StatusInternalServerError)
		return
	}
	log.Printf("Updated loyalty points for user %d by %d points", req.UserID, points)

	// Log the transaction action
	utils.LogAction(db, req.UserID, "Add Transaction", fmt.Sprintf("Transaction ID %s recorded with %d points", req.TransactionID, points))

	// Respond with success message and points
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AddTransactionResponse{
		Message: "Transaction recorded successfully",
		Points:  points,
	})
}
