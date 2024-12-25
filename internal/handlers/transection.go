package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"loyalty-points-system-api/internal/models"
	response "loyalty-points-system-api/internal/reponse"
	utils "loyalty-points-system-api/internal/utils"
	"loyalty-points-system-api/pkg/middleware"
	"net/http"
	"time"
)

// AddTransactionHandler - Adds transaction and updates points consistently
func AddTransactionHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("AddTransactionHandler: Starting to process add transaction request.")

	// Extract the username from the token (context)
	tokenUsername, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		response.WriteErrorResponse(w, http.StatusUnauthorized, response.APIError{
			Code:    "401",
			Msg:     "Unauthorized",
			Details: "Failed to extract user information from token",
		})
		return
	}

	// Parse the request body
	var req models.AddTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteErrorResponse(w, http.StatusBadRequest, response.APIError{
			Code:    "400",
			Msg:     "Invalid Request Payload",
			Details: "Failed to decode JSON body",
		})
		return
	}

	// Validate that the user_id in the request matches the logged-in user
	var dbUsername string
	err := db.QueryRow("SELECT username FROM users WHERE id = ?", req.UserID).Scan(&dbUsername)
	if err == sql.ErrNoRows {
		response.WriteErrorResponse(w, http.StatusNotFound, response.APIError{
			Code:    "404",
			Msg:     "User Not Found",
			Details: "User ID does not exist",
		})
		return
	} else if err != nil {
		log.Printf("Error fetching user data: %v", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Failed to fetch user data",
		})
		return
	}

	if tokenUsername != dbUsername {
		response.WriteErrorResponse(w, http.StatusForbidden, response.APIError{
			Code:    "403",
			Msg:     "Forbidden",
			Details: "You can only create transactions for your own account",
		})
		return
	}

	// Calculate points based on category
	categoryMultipliers := map[string]float64{
		"electronics": 1.0,
		"groceries":   2.0,
		"clothing":    1.5,
	}

	multiplier, ok := categoryMultipliers[req.Category]
	if !ok {
		response.WriteErrorResponse(w, http.StatusBadRequest, response.APIError{
			Code:    "400",
			Msg:     "Invalid Category",
			Details: "The category provided is not valid",
		})
		return
	}

	pointsEarned := int(req.TransactionAmount * multiplier)
	log.Printf("Calculated %d points for user %d in category %s", pointsEarned, req.UserID, req.Category)

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Transaction Error",
			Details: "Failed to start database transaction",
		})
		return
	}
	defer tx.Rollback()

	// Record the transaction
	_, err = tx.Exec(`
		INSERT INTO transactions (
			transaction_id, user_id, transaction_amount, 
			category, transaction_date, product_code, points
		) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		req.TransactionID, req.UserID, req.TransactionAmount,
		req.Category, req.TransactionDate, req.ProductCode, pointsEarned,
	)
	if err != nil {
		log.Printf("Error recording transaction: %v", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Transaction Error",
			Details: "Could not record transaction",
		})
		return
	}

	// Add points record
	validUntil := time.Now().AddDate(1, 0, 0) // Points valid for 1 year
	_, err = tx.Exec(`
		INSERT INTO points (
			user_id, transaction_id, points, 
			transaction_type, transaction_date, valid_until, reason
		) VALUES (?, ?, ?, 'Earned', ?, ?, ?)`,
		req.UserID, req.TransactionID, pointsEarned,
		req.TransactionDate, validUntil, "Purchase",
	)
	if err != nil {
		log.Printf("Error recording points: %v", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Transaction Error",
			Details: "Could not record points",
		})
		return
	}

	// Update user's total loyalty points
	_, err = tx.Exec(`
		UPDATE users 
		SET loyalty_points = loyalty_points + ? 
		WHERE id = ?`,
		pointsEarned, req.UserID,
	)
	if err != nil {
		log.Printf("Error updating user points: %v", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Transaction Error",
			Details: "Could not update user points",
		})
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Transaction Error",
			Details: "Could not commit transaction",
		})
		return
	}

	// Get updated balance
	var currentPoints int
	err = db.QueryRow("SELECT loyalty_points FROM users WHERE id = ?", req.UserID).Scan(&currentPoints)
	if err != nil {
		log.Printf("Error fetching updated points balance: %v", err)
	}

	utils.LogAction(db, req.UserID, "Add Transaction",
		fmt.Sprintf("Transaction %s: Earned %d points. New balance: %d",
			req.TransactionID, pointsEarned, currentPoints))

	response.WriteSuccessResponse(w, models.AddTransactionResponse{
		Message: "Transaction recorded successfully",
		Points:  pointsEarned,
	}, "Transaction recorded successfully")
}
