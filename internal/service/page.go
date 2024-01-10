package service

import (
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/repository"
)

type PageService struct {
	repo repository.Page
}

func NewPageService(repo repository.Page) *PageService {
	return &PageService{repo: repo}
}

func (s *PageService) GetByID(id int64) (domain.Page, error) {
	page, err := s.repo.GetByID(id)
	return page, err
}

func (s *PageService) GetChildrens(id int64) ([]domain.Page, error) {
	childs, err := s.repo.GetChildrens(id)
	return childs, err
}

func (s *PageService) GetDefaultPage(pageID int64) ([]domain.DefaultPage, error) {
	pages, err := s.repo.GetDefaultPage(pageID)
	return pages, err
}

func (s *PageService) GetRootPages(organizationID *int64) ([]domain.Page, error) {
	pages, err := s.repo.GetRootPages(organizationID)
	return pages, err
}

func (s *PageService) GetPages(organizationID *int64) ([]domain.Page, error) {
	pages, err := s.repo.GetPages(organizationID)
	return pages, err
}

func (s *PageService) Create(organizationID *int64, parentID int64, languageTag, pageType, theme, status, title string, image, imageHover *string) (int64, error) {
	id, err := s.repo.Create(organizationID, parentID, languageTag, pageType, theme, status, title, image, imageHover)
	return id, err
}

func (s *PageService) CreateDefaultPage(pageID int64, plans []string) error {
	err := s.repo.CreateDefaultPage(pageID, plans)
	return err
}

func (s *PageService) Update(id int64, parentID *int64, status, title, image, imageHover *string) error {
	err := s.repo.Update(id, parentID, status, title, image, imageHover)
	return err
}

func (s *PageService) UpdateSort(parentID int64, sort []int) error {
	err := s.repo.UpdateSort(parentID, sort)
	return err
}

func (s *PageService) UpdateUpdatedAt(pageID int64) error {
	err := s.repo.UpdateUpdatedAt(pageID)
	return err
}

func (s *PageService) Delete(id int64) error {
	err := s.repo.Delete(id)
	return err
}

func (s *PageService) DeleteDefaultPage(pageID int64, plans []string) error {
	err := s.repo.DeleteDefaultPage(pageID, plans)
	return err
}
