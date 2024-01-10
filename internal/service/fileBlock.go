package service

import (
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/repository"
)

type FileBlockService struct {
	repo repository.FileBlock
}

func NewFileBlockService(repo repository.FileBlock) *FileBlockService {
	return &FileBlockService{repo: repo}
}

func (s *FileBlockService) GetByID(id int64) (domain.FileBlock, error) {
	file, err := s.repo.GetByID(id)
	return file, err
}

func (s *FileBlockService) GetByPageID(pageID int64) (domain.FileBlock, error) {
	file, err := s.repo.GetByPageID(pageID)
	return file, err
}

func (s *FileBlockService) Create(pageID int64, path string) (int64, error) {
	id, err := s.repo.Create(pageID, path)
	return id, err
}

func (s *FileBlockService) Update(id int64, path string) error {
	err := s.repo.Update(id, path)
	return err
}
