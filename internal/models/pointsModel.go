package models

import "time"

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
