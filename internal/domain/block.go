package domain

import "gopkg.in/guregu/null.v4"

type Block struct {
	ID         int64       `db:"id" json:"id"`
	PageID     int64       `db:"page_id" json:"pageID"`
	Title      string      `db:"title" json:"title"`
	Content    null.String `db:"content" json:"content"`
	Readmore   null.String `db:"readmore" json:"readmore"`
	Image      null.String `db:"image" json:"image"`
	ImageHover null.String `db:"image_hover" json:"imageHover"`
	Type       string      `db:"type" json:"type"`
	Sort       int         `db:"sort" json:"sort"`
}
