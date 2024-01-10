package domain

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type User struct {
	ID         int64       `db:"id" json:"id"`
	Firstname  null.String `db:"firstname" json:"firstname"`
	Lastname   null.String `db:"lastname" json:"lastname"`
	Salutation null.String `db:"salutation" json:"salutation"`
	Email      string      `db:"email" json:"email"`
	Image      null.String `db:"image" json:"image"`
	Password   null.String `db:"password" json:"password"`
	Type       string      `db:"type" json:"type"`
	CreatedAt  time.Time   `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time   `db:"updated_at" json:"updatedAt"`
}

type UserVerification struct {
	ID     int64  `db:"id" json:"id"`
	Token  string `db:"token" json:"token"`
	Status string `db:"status" json:"status"`
}

type UserPasswordReset struct {
	ID    int64  `db:"id" json:"id"`
	Token string `db:"token" json:"token"`
}
