package repository

import (
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

type FileRepo struct {
	db       *sqlx.DB
	pageRepo *PageRepo
}

func NewFileRepo(db *sqlx.DB, pageRepo *PageRepo) *FileRepo {
	return &FileRepo{db: db, pageRepo: pageRepo}
}

func (r *FileRepo) GetByID(id int64) (domain.FileBlock, error) {
	var file domain.FileBlock
	err := r.db.Get(&file, "SELECT * FROM files WHERE id = ?", id)
	return file, err
}

func (r *FileRepo) GetByPageID(pageID int64) (domain.FileBlock, error) {
	var file domain.FileBlock
	err := r.db.Get(&file, "SELECT * FROM files WHERE id = ?", pageID)
	return file, err
}

func (r *FileRepo) Create(pageID int64, path string) (int64, error) {
	result, err := r.db.Exec(`INSERT INTO files (id, path) VALUES (?, ?) ON DUPLICATE KEY UPDATE path = ?`, pageID, path, path)
	if err != nil {
		return 0, err
	}
	fileID, _ := result.LastInsertId()

	go r.pageRepo.UpdateUpdatedAt(pageID)

	return fileID, nil
}

func (r *FileRepo) Update(id int64, path string) error {
	_, err := r.db.Exec("UPDATE files SET path = ? WHERE id = ?", path, id)
	if err != nil {
		return err
	}

	go r.pageRepo.UpdateUpdatedAt(id)

	return nil
}
