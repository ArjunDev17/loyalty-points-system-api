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

type RedeemRequest struct {
	UserID int `json:"user_id"`
	Points int `json:"points"`
}

type RedeemResponse struct {
	Success         bool   `json:"success"`
	Message         string `json:"message"`
	RemainingPoints int    `json:"remaining_points"`
}

type PointsHistory struct {
	TransactionDate string `json:"transaction_date"`
	Points          int    `json:"points"`
	Reason          string `json:"reason"`
}

type PointsBalanceResponse struct {
	Balance int             `json:"balance"`
	History []PointsHistory `json:"history"`
}
