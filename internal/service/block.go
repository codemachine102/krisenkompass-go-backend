package service

import (
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/repository"
)

type BlockService struct {
	repo repository.Block
}

func NewBlockService(repo repository.Block) *BlockService {
	return &BlockService{repo: repo}
}

func (s *BlockService) GetByPageID(pageID int64) ([]domain.Block, error) {
	blocks, err := s.repo.GetByPageID(pageID)
	return blocks, err
}

func (s *BlockService) GetByID(id int64) (domain.Block, error) {
	block, err := s.repo.GetByID(id)
	return block, err
}

func (s *BlockService) Create(pageID int64, title, blockType string, content, readmore, image, imageHover *string) (int64, error) {
	id, err := s.repo.Create(pageID, title, blockType, content, readmore, image, imageHover)
	return id, err
}

func (s *BlockService) Update(id int64, title, content, readmore, image, imageHover *string) error {
	err := s.repo.Update(id, title, content, readmore, image, imageHover)
	return err
}

func (s *BlockService) UpdateSort(pageID int64, sort []int) error {
	err := s.repo.UpdateSort(pageID, sort)
	return err
}

func (s *BlockService) Delete(id int64) error {
	err := s.repo.Delete(id)
	return err
}
