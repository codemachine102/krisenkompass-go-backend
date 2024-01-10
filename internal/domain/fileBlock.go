package domain

import "time"

type FileBlock struct {
	ID        int64     `db:"id" json:"id"`
	Path      string    `db:"path" json:"path"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}
