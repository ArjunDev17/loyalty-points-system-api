package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"loyalty-points-system-api/internal/models"
	response "loyalty-points-system-api/internal/reponse"
	"net/http"
)

// PointsHistoryHandler handles retrieving a user's points history
func PointsHistoryHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("PointsHistoryHandler: Starting to process points history request.")

	var req models.PointsHistoryRequest
	if r.Body == nil {
		log.Println("PointsHistoryHandler: Request body is empty.")
		response.WriteErrorResponse(w, http.StatusBadRequest, response.APIError{
			Code:    "400",
			Msg:     "Empty Request Body",
			Details: "Request body cannot be empty",
		})
		return
	}

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		response.WriteErrorResponse(w, http.StatusBadRequest, response.APIError{
			Code:    "400",
			Msg:     "Invalid Request Payload",
			Details: "Failed to decode JSON body",
		})
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
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Failed to fetch points history",
		})
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
			response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
				Code:    "500",
				Msg:     "Internal Server Error",
				Details: "Failed to process points history",
			})
			return
		}
		history = append(history, record)
	}

	// Check for errors during row iteration
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v", err)
		response.WriteErrorResponse(w, http.StatusInternalServerError, response.APIError{
			Code:    "500",
			Msg:     "Internal Server Error",
			Details: "Failed to process points history",
		})
		return
	}

	// Respond with the history
	log.Printf("PointsHistoryHandler: Successfully retrieved %d records for user_id: %d", len(history), req.UserID)
	response.WriteSuccessResponse(w, history, "Points history retrieved successfully")
}
