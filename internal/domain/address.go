package domain

import "gopkg.in/guregu/null.v4"

type Address struct {
	ID             int64       `db:"id" json:"id"`
	OrganizationID int64       `db:"organization_id" json:"organizationID"`
	Firstname      string      `db:"firstname" json:"firstname"`
	Lastname       string      `db:"lastname" json:"lastname"`
	Email          string      `db:"email" json:"email"`
	Phone          string      `db:"phone" json:"phone"`
	PhoneExtra     null.String `db:"phone_extra" json:"phoneExtra"`
	Role           null.String `db:"role" json:"role"`
	Info           null.String `db:"info" json:"info"`
	Sort           int         `db:"sort" json:"sort"`
}
