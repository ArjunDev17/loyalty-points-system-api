package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"loyalty-points-system-api/internal/models"
	response "loyalty-points-system-api/internal/reponse"
	utils "loyalty-points-system-api/internal/utils"
	"net/http"
	"time"
)

// RedeemPointsHandler - Redeems points and updates both tables
func RedeemPointsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("RedeemPointsHandler: Starting to process redeem points request.")

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

	tx, err := db.Begin()
	if err != nil {
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
	err = tx.QueryRow(`
		SELECT loyalty_points 
		FROM users 
		WHERE id = ? 
		FOR UPDATE`, req.UserID).Scan(&totalPoints)
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

	// First create a transaction record for the redemption
	_, err = tx.Exec(`
		INSERT INTO transactions (
			transaction_id, user_id, transaction_amount,
			category, transaction_date, product_code, points
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

	// Find points to redeem from points table
	pointsToRedeem := req.Points
	rows, err := tx.Query(`
		SELECT id, points 
		FROM points 
		WHERE user_id = ? AND valid_until > NOW() 
		ORDER BY valid_until ASC 
		FOR UPDATE`, req.UserID)
	if err != nil {
		log.Printf("Error fetching points records: %v", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Failed to fetch points records",
		})
		return
	}
	defer rows.Close()

	// Process points deduction
	var pointsRecords []struct {
		ID     int
		Points int
	}
	for rows.Next() {
		var record struct {
			ID     int
			Points int
		}
		if err := rows.Scan(&record.ID, &record.Points); err != nil {
			log.Printf("Error scanning points record: %v", err)
			response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
				Code:    "500",
				Msg:     "Internal Server Error",
				Details: "Failed to process points records",
			})
			return
		}
		pointsRecords = append(pointsRecords, record)
	}

	// Update or delete points records
	for _, record := range pointsRecords {
		if pointsToRedeem <= 0 {
			break
		}

		if record.Points <= pointsToRedeem {
			_, err := tx.Exec("DELETE FROM points WHERE id = ?", record.ID)
			if err != nil {
				log.Printf("Error deleting points record: %v", err)
				response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
					Code:    "500",
					Msg:     "Internal Server Error",
					Details: "Failed to update points record",
				})
				return
			}
			pointsToRedeem -= record.Points
		} else {
			_, err := tx.Exec("UPDATE points SET points = points - ? WHERE id = ?",
				pointsToRedeem, record.ID)
			if err != nil {
				log.Printf("Error updating points record: %v", err)
				response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
					Code:    "500",
					Msg:     "Internal Server Error",
					Details: "Failed to update points record",
				})
				return
			}
			pointsToRedeem = 0
		}
	}

	// Update user's total points
	_, err = tx.Exec("UPDATE users SET loyalty_points = loyalty_points - ? WHERE id = ?",
		req.Points, req.UserID)
	if err != nil {
		log.Printf("Error updating user points: %v", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Failed to update user points",
		})
		return
	}

	// Record redemption in points table
	_, err = tx.Exec(`
		INSERT INTO points (
			user_id, transaction_id, points, 
			transaction_type, transaction_date, 
			valid_until, reason
		) VALUES (?, ?, ?, 'Redeemed', NOW(), NOW(), ?)`,
		req.UserID,
		redemptionTxnID,
		-req.Points,
		fmt.Sprintf("Points Redemption (ID: %s)", redemptionTxnID))
	if err != nil {
		log.Printf("Error recording redemption: %v", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Failed to record redemption",
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

	// Get final balance
	var remainingPoints int
	err = db.QueryRow("SELECT loyalty_points FROM users WHERE id = ?",
		req.UserID).Scan(&remainingPoints)
	if err != nil {
		log.Printf("Error fetching final balance: %v", err)
	}

	utils.LogAction(db, req.UserID, "Redeem Points",
		fmt.Sprintf("Redeemed %d points (Transaction ID: %s). New balance: %d",
			req.Points, redemptionTxnID, remainingPoints))

	response.WriteSuccessResponse(w, map[string]interface{}{
		"remaining_points": remainingPoints,
		"points_redeemed":  req.Points,
		"redemption_id":    redemptionTxnID,
	}, "Points redeemed successfully")
}
