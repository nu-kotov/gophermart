package models

import "time"

type UserData struct {
	UserID   string `json:"user_id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type OrderData struct {
	Number     int64     `json:"number"`
	UserID     string    `json:"user_id"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type UserBalance struct {
	Balance   float64 `json:"balance"`
	Withdrawn float64 `json:"withdrawn"`
}

type GetUserOrdersResponse struct {
	Number     int64   `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual"`
	UploadedAt string  `json:"uploaded_at"`
}

type WithdrawnInfo struct {
	Number      string  `json:"order"`
	Sum         float64 `json:"sum"`
	WithdrawnAt string  `json:"withdrawn_at"`
}

type Withdraw struct {
	Number      int64     `json:"order"`
	UserID      string    `json:"user_id"`
	Sum         float64   `json:"sum"`
	WithdrawnAt time.Time `json:"withdrawn_at"`
}
