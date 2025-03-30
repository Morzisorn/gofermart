package models

import "time"

type User struct {
	Login     string   `json:"login"`
	Password  [32]byte `json:-`
	Current   float64  `json:"current"`
	Withdrawn float64  `json:"withdrawn"`
}

type ParseUserRegister struct {
	Login     string   `json:"login"`
	Password  string `json:"password"`
}

type Order struct {
	Number     string    `json:"number"`
	UploadedAt time.Time `json:"uploaded_at"`
	UserLogin  string    `json:"user_login,omitempty"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
}


type UserBalance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Withdrawal struct {
	Number      string    `json:"order"`
	ProcessedAt time.Time `json:"processed_at"`
	UserLogin   string    `json:"user_login,omitempty"`
	Sum         float64   `json:"sum"`
}

type LoyaltyOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}



const (
	OrderStatusNEW        string = "NEW"
	OrderStatusPROCESSING string = "PROCESSING"
	OrderStatusINVALID    string = "INVALID"
	OrderStatusPROCESSED  string = "PROCESSED"
)

const (
	LoyaltyStatusREGISTERED string = "REGISTERED"
	LoyaltyStatusINVALID    string = "INVALID"
	LoyaltyStatusPROCESSING string = "PROCESSING"
	LoyaltyStatusPROCESSED  string = "PROCESSED"
)
