package service

import (
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/repository"
	"gopkg.in/guregu/null.v4"
)

type AddressService struct {
	repo repository.Address
}

func NewAddressService(repo repository.Address) *AddressService {
	return &AddressService{repo: repo}
}

func (s *AddressService) GetByID(id int64) (domain.Address, error) {
	address, err := s.repo.GetByID(id)
	return address, err
}

func (s *AddressService) GetByOrganizationID(organizationID int64) ([]domain.Address, error) {
	addresses, err := s.repo.GetByOrganizationID(organizationID)
	return addresses, err
}

func (s *AddressService) Create(organizationID int64, firstname, lastname, email, phone string, phoneExtra, role, info null.String) (int64, error) {
	id, err := s.repo.Create(organizationID, firstname, lastname, email, phone, phoneExtra, role, info)
	return id, err
}

func (s *AddressService) Update(id int64, firstname, lastname, email, phone, phoneExtra, role, info null.String) error {
	err := s.repo.Update(id, firstname, lastname, email, phone, phoneExtra, role, info)
	return err
}

func (s *AddressService) UpdateSort(organizationID int64, sort []int64) error {
	err := s.repo.UpdateSort(organizationID, sort)
	return err
}

func (s *AddressService) Delete(id int64) error {
	err := s.repo.Delete(id)
	return err
}
