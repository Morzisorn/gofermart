// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package database

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Order struct {
	Number     string           `json:"number"`
	UploadedAt pgtype.Timestamp `json:"uploaded_at"`
	UserLogin  string           `json:"user_login"`
	Status     pgtype.Text      `json:"status"`
	Accrual    pgtype.Float4    `json:"accrual"`
}

type User struct {
	Login     string        `json:"login"`
	Password  []byte        `json:"password"`
	Current   pgtype.Float4 `json:"current"`
	Withdrawn pgtype.Float4 `json:"withdrawn"`
}

type Withdrawal struct {
	Number      string           `json:"number"`
	ProcessedAt pgtype.Timestamp `json:"processed_at"`
	UserLogin   string           `json:"user_login"`
	Sum         pgtype.Float4    `json:"sum"`
}
