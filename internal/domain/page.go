package domain

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type Page struct {
	ID             int64       `db:"id" json:"id"`
	OrganizationID null.Int    `db:"organization_id" json:"organizationID"`
	ParentID       null.Int    `db:"parent_id" json:"parentID"`
	LanguageTag    string      `db:"language_tag" json:"languageTag"`
	Type           string      `db:"type" json:"type"`
	Theme          string      `db:"theme" json:"theme"`
	Status         string      `db:"status" json:"status"`
	Title          string      `db:"title" json:"title"`
	Image          null.String `db:"image" json:"image"`
	ImageHover     null.String `db:"image_hover" json:"imageHover"`
	Sort           int         `db:"sort" json:"sort"`
	CreatedAt      time.Time   `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time   `db:"updated_at" json:"updatedAt"`
}
