package service

import (
	"database/sql"
	"errors"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/repository"
	pass "bitbucket.org/ibros_nsk/krisenkompass-backend/pkg/password"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/pkg/randstring"
)

type UserService struct {
	repo                repository.User
	organizationService Organization
	emailService        Email
}

func NewUserService(repo repository.User, organizationService Organization, emailService Email) *UserService {
	return &UserService{repo: repo, organizationService: organizationService, emailService: emailService}
}

func (s *UserService) GetAll() ([]domain.User, error) {
	organizations, err := s.repo.GetAll()
	return organizations, err
}

func (s *UserService) GetByID(id int64) (domain.User, error) {
	user, err := s.repo.GetByID(id)
	return user, err
}

func (s *UserService) GetByEmail(email string) (domain.User, error) {
	user, err := s.repo.GetByEmail(email)
	return user, err
}

func (s *UserService) GetByPasswordResetToken(token string) (domain.User, error) {
	user, err := s.repo.GetByPasswordResetToken(token)
	return user, err
}

func (s *UserService) GetByOrganizationID(organizationID int64) ([]domain.User, error) {
	users, err := s.repo.GetByOrganizationID(organizationID)
	return users, err
}

func (s *UserService) GetVerification(id int64) (domain.UserVerification, error) {
	verification, err := s.repo.GetVerification(id)
	return verification, err
}

func (s *UserService) GetPasswordReset(email string) (domain.UserPasswordReset, error) {
	reset, err := s.repo.GetPasswordReset(email)
	return reset, err
}

func (s *UserService) Create(firstname, lastname *string, email string, image *string) (int64, error) {
	userID, err := s.repo.Create(firstname, lastname, email, image)
	if err != nil {
		return 0, err
	}

	token := randstring.RandAlphanumString(32)
	err = s.repo.CreateVerification(userID, token)
	if err != nil {
		return 0, err
	}

	err = s.emailService.SendUserVerificatonLink(email, token)
	if err != nil {
		// Rollback user creation
		s.Delete(userID)
		return 0, err
	}

	return userID, nil
}

func (s *UserService) CreateWithOrganization(firstname, lastname *string, email string, name, city, address, invoiceAddress, plan string, population int, role string) (userID int64, organizationID int64, err error) {
	userID, organizationID, err = s.repo.CreateWithOrganization(firstname, lastname, email, name, city, address, invoiceAddress, plan, population)
	if err != nil {
		return 0, 0, err
	}

	token := randstring.RandAlphanumString(32)
	err = s.repo.CreateVerification(userID, token)
	if err != nil {
		return 0, 0, err
	}

	err = s.emailService.SendUserVerificatonLink(email, token)
	if err != nil {
		// Rollback
		s.Delete(userID)
		s.organizationService.Delete(organizationID)
		return 0, 0, err
	}

	return userID, organizationID, nil
}

func (s *UserService) CreateOrganizationUser(organizationID int64, email, role string) error {
	isNewUser := false
	token := ""

	organization, err := s.organizationService.GetByID(organizationID)
	if err != nil {
		return errors.New("failed to find organization")
	}

	user, err := s.GetByEmail(email)
	if err != nil && err != sql.ErrNoRows {
		return errors.New("failed to invite user")
	} else if err == sql.ErrNoRows {
		// If user doesn't exist then create new one.
		isNewUser = true

		userID, err := s.repo.Create(nil, nil, email, nil)
		if err != nil {
			return errors.New("failed to invite user")
		}

		token = randstring.RandAlphanumString(32)
		err = s.repo.CreateVerification(userID, token)
		if err != nil {
			// Rollback user creation
			go s.Delete(userID)
			return err
		}

		user, err = s.GetByID(userID)
		if err != nil {
			return errors.New("failed to invite user")
		}
	}

	// Add user to the organization.
	err = s.repo.CreateOrganizationUser(organizationID, user.ID, role)
	if err != nil {
		if isNewUser {
			// Rollback user creation
			go s.Delete(user.ID)
		}
		return errors.New("failed to invite user")
	}

	// Send email
	if isNewUser {
		err = s.emailService.SendUserVerificatonLinkWithInvite(email, token, organization.Name)
		if err != nil {
			// Rollback user creation
			go s.Delete(user.ID)
			return err
		}
	} else {
		err = s.emailService.SendUserInvite(email, organizationID, organization.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *UserService) CreatePasswordReset(email string) error {
	token := randstring.RandAlphanumString(32)
	err := s.repo.CreatePasswordReset(email, token)
	if err != nil {
		return err
	}

	err = s.emailService.SendPasswordResetLink(email, token)
	if err != nil {
		return err
	}

	return err
}

func (s *UserService) Update(id int64, image, firstname, lastname, email, userType *string) error {
	err := s.repo.Update(id, image, firstname, lastname, email, userType)

	return err
}

func (s *UserService) UpdatePassword(id int64, password string) error {
	hashedPassword, err := pass.HashPassword(password)
	if err != nil {
		return err
	}

	err = s.repo.UpdatePassword(id, hashedPassword)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) UpdateOrganizationUser(organizationID, userID int64, role string) error {
	err := s.repo.UpdateOrganizationUser(organizationID, userID, role)

	return err
}

func (s *UserService) Delete(id int64) error {
	err := s.repo.Delete(id)

	return err
}

func (s *UserService) DeleteOrganizationUser(organizationID, userID int64) error {
	err := s.repo.DeleteOrganizationUser(organizationID, userID)

	return err
}
