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

// RedeemPointsHandler - Redeems points and updates both tables
func RedeemPointsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("RedeemPointsHandler: Starting to process redeem points request.")

	// Get the username from the token (context)
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
	var req models.RedeemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		response.WriteErrorResponse(w, http.StatusBadRequest, response.APIError{
			Code:    "400",
			Msg:     "Invalid Request Body",
			Details: "Failed to decode JSON body",
		})
		return
	}

	// Validate that the username from the token matches the user_id in the request
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
			Details: "You can only redeem your own points",
		})
		return
	}

	// Proceed with the redeem points logic
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Transaction start error: %v", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Failed to start transaction",
		})
		return
	}
	defer tx.Rollback()

	// Check available points
	var totalPoints int
	err = tx.QueryRow("SELECT loyalty_points FROM users WHERE id = ? FOR UPDATE", req.UserID).Scan(&totalPoints)
	if err != nil {
		log.Printf("Error fetching user points: %v", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Failed to fetch user points",
		})
		return
	}

	if req.Points > totalPoints {
		response.WriteErrorResponse(w, http.StatusBadRequest, response.APIError{
			Code:    "400",
			Msg:     "Insufficient Points",
			Details: "User does not have enough points for redemption",
		})
		return
	}

	// Generate redemption transaction ID
	redemptionTxnID := fmt.Sprintf("RED_%d_%s", req.UserID, time.Now().Format("20060102150405"))

	// Create a transaction record for the redemption
	_, err = tx.Exec(`
		INSERT INTO transactions (
			transaction_id, user_id, transaction_amount, category, transaction_date, product_code, points
		) VALUES (?, ?, ?, 'redemption', NOW(), 'REDEMPTION', ?)`,
		redemptionTxnID, req.UserID, 0, -req.Points)
	if err != nil {
		log.Printf("Error creating redemption transaction: %v", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Failed to create redemption transaction",
		})
		return
	}

	// Deduct points and commit the transaction
	_, err = tx.Exec("UPDATE users SET loyalty_points = loyalty_points - ? WHERE id = ?", req.Points, req.UserID)
	if err != nil {
		log.Printf("Error updating user points: %v", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Failed to update user points",
		})
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Failed to commit transaction",
		})
		return
	}

	// Fetch the updated points balance
	var remainingPoints int
	err = db.QueryRow("SELECT loyalty_points FROM users WHERE id = ?", req.UserID).Scan(&remainingPoints)
	if err != nil {
		log.Printf("Error fetching final balance: %v", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Failed to fetch final points balance",
		})
		return
	}

	utils.LogAction(db, req.UserID, "Redeem Points", fmt.Sprintf("Redeemed %d points. Transaction ID: %s", req.Points, redemptionTxnID))

	// Respond with the final balance and redemption details
	response.WriteSuccessResponse(w, map[string]interface{}{
		"remaining_points": remainingPoints,
		"points_redeemed":  req.Points,
		"redemption_id":    redemptionTxnID,
	}, "Points redeemed successfully")
}
