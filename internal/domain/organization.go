package domain

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type Organization struct {
	ID             int64       `db:"id" json:"id"`
	Name           string      `db:"name" json:"name"`
	Image          null.String `db:"image" json:"image"`
	City           string      `db:"city" json:"city"`
	Street         string      `json:"street"`
	Population     int         `db:"population" json:"population"`
	Address        string      `db:"address" json:"address"`
	InvoiceAddress string      `db:"invoice_address" json:"invoiceAddress"`
	Plan           string      `db:"plan" json:"plan"`
	Status         string      `db:"status" json:"status"`
	CreatedAt      time.Time   `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time   `db:"updated_at" json:"updatedAt"`
}
