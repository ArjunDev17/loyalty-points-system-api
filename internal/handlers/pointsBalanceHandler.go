package handlers

import (
	"database/sql"
	"log"
	"loyalty-points-system-api/internal/models"
	response "loyalty-points-system-api/internal/reponse"
	"net/http"

	"strconv"
)

// PointsBalanceHandler returns the user's current points balance and history
func PointsBalanceHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("PointsBalanceHandler: Starting to process points balance request.")

	// Parse query parameters for pagination
	userID := r.URL.Query().Get("user_id")
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	if userID == "" {
		log.Println("PointsBalanceHandler: user_id is missing in the query parameters.")
		response.WriteErrorResponse(w, http.StatusBadRequest, response.APIError{
			Code:    "400",
			Msg:     "Missing Parameter",
			Details: "user_id is required",
		})
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
		log.Printf("Error retrieving points balance for user %s: %v", userID, err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Could not retrieve points balance",
		})
		return
	}

	// Get the user's points history
	history := []models.PointsHistory{}
	query := `SELECT transaction_date, points, category FROM transactions WHERE user_id = ? ORDER BY transaction_date DESC LIMIT ? OFFSET ?`
	rows, err := db.Query(query, userID, pageSize, offset)
	if err != nil {
		log.Printf("Error retrieving points history for user %s: %v", userID, err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Could not retrieve points history",
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var record models.PointsHistory
		var category string
		if err := rows.Scan(&record.TransactionDate, &record.Points, &category); err != nil {
			log.Printf("Error scanning points history row for user %s: %v", userID, err)
			response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
				Code:    "500",
				Msg:     "Internal Server Error",
				Details: "Error scanning points history",
			})
			return
		}
		record.Reason = "Transaction - " + category
		history = append(history, record)
	}

	// Check for errors during row iteration
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows for user %s: %v", userID, err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Failed to process points history",
		})
		return
	}

	// Respond with balance and history
	responses := models.PointsBalanceResponse{
		Balance: balance,
		History: history,
	}
	log.Printf("PointsBalanceHandler: Successfully retrieved points balance and history for user %s.", userID)
	response.WriteSuccessResponse(w, responses, "Points balance and history retrieved successfully")
}
