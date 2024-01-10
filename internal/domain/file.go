package domain

import "time"

type File struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	IsDir     bool      `json:"isDir"`
	UpdatedAt time.Time `json:"updatedAt"`
}
