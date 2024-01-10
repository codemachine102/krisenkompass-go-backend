package service

import (
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/repository"
	"gopkg.in/guregu/null.v4"
)

type OrganizationService struct {
	repo repository.Organization
}

func NewOrganizationService(organizationRepo repository.Organization) *OrganizationService {
	return &OrganizationService{repo: organizationRepo}
}

func (s *OrganizationService) GetAll() ([]domain.Organization, error) {
	organizations, err := s.repo.GetAll()
	return organizations, err
}

func (s *OrganizationService) GetByID(id int64) (domain.Organization, error) {
	organization, err := s.repo.GetByID(id)
	return organization, err
}

func (s *OrganizationService) GetByUserID(userID int64) ([]domain.Organization, error) {
	organizations, err := s.repo.GetByUserID(userID)
	return organizations, err
}

func (s *OrganizationService) Create(image *string, name, city, address, invoiceAddress, plan string, population int, userID int64) (int64, error) {
	id, err := s.repo.Create(image, name, city, address, invoiceAddress, plan, population, userID)
	return id, err
}

func (s *OrganizationService) CopyDefaultContent(organizationID int64, plan string) error {
	err := s.repo.CopyDefaultContent(organizationID, plan)
	return err
}

func (s *OrganizationService) Update(id int64, name, image, city, address, invoiceAddress, plan, status null.String, population null.Int) error {
	err := s.repo.Update(id, image, name, city, address, invoiceAddress, plan, status, population)
	return err
}

func (s *OrganizationService) Delete(id int64) error {
	err := s.repo.Delete(id)
	return err
}
