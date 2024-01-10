package service

import (
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/repository"
)

type OrganizationUserService struct {
	repo repository.OrganizationUser
}

func NewOrganizationUserService(repo repository.OrganizationUser) *OrganizationUserService {
	return &OrganizationUserService{repo: repo}
}

func (s *OrganizationUserService) GetByUserID(id int64) ([]domain.OrganizationUser, error) {
	rows, err := s.repo.GetByUserID(id)
	return rows, err
}
