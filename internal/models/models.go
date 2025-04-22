package models

import (
	"time"
)

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
