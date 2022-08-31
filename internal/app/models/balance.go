package models

type UserBalance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Transaction struct {
	UserID      int        `json:"-"`
	OrderNum    string     `json:"order"`
	Amount      float64    `json:"sum"`
	ProcessedAt CustomTime `json:"processed_at"`
}
