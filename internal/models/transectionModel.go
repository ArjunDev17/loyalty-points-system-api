package models

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
