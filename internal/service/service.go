package service

import (
	"io/fs"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/repository"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/pkg/cache"
	"gopkg.in/guregu/null.v4"
)

type Auth interface {
	Login(email, password string) (domain.Tokens, error)
	RefreshToken(refreshToken string) (domain.Tokens, error)
	Verify(token, password string) (domain.Tokens, error)
}

type User interface {
	GetAll() ([]domain.User, error)
	GetByID(id int64) (domain.User, error)
	GetByEmail(email string) (domain.User, error)
	GetVerification(id int64) (domain.UserVerification, error)
	GetPasswordReset(email string) (domain.UserPasswordReset, error)
	GetByPasswordResetToken(token string) (domain.User, error)
	GetByOrganizationID(organizationID int64) ([]domain.User, error)
	Create(firstname, lastname *string, email string, image *string) (int64, error)
	CreateWithOrganization(firstname, lastname *string, email string, name, city, address, invoiceAddress, plan string, population int, role string) (userID int64, organizationID int64, err error)
	CreateOrganizationUser(organizationID int64, email, role string) error
	CreatePasswordReset(email string) error
	Update(id int64, image, firstname, lastname, email, userType *string) error
	UpdatePassword(id int64, password string) error
	UpdateOrganizationUser(organizationID, userID int64, role string) error
	Delete(id int64) error
	DeleteOrganizationUser(organizationID, userID int64) error
}

type Organization interface {
	GetAll() ([]domain.Organization, error)
	GetByID(id int64) (domain.Organization, error)
	GetByUserID(userID int64) ([]domain.Organization, error)
	Create(image *string, name, city, address, invoiceAddress, plan string, population int, userID int64) (int64, error)
	Update(id int64, image, name, city, address, invoiceAddress, plan, status null.String, population null.Int) error
	Delete(id int64) error
	CopyDefaultContent(organizationID int64, plan string) error
}

type OrganizationUser interface {
	GetByUserID(id int64) ([]domain.OrganizationUser, error)
}

type Page interface {
	GetByID(id int64) (domain.Page, error)
	GetChildrens(id int64) ([]domain.Page, error)
	GetDefaultPage(pageID int64) ([]domain.DefaultPage, error)
	GetRootPages(organizationID *int64) ([]domain.Page, error)
	GetPages(organizationID *int64) ([]domain.Page, error)
	Create(organizationID *int64, parentID int64, languageTag, pageType, theme, status, title string, image, imageHover *string) (int64, error)
	CreateDefaultPage(pageID int64, plans []string) error
	Update(id int64, parentID *int64, status, title, image, imageHover *string) error
	UpdateSort(parentID int64, sort []int) error
	UpdateUpdatedAt(pageID int64) error
	Delete(id int64) error
	DeleteDefaultPage(pageID int64, plans []string) error
}

type Block interface {
	GetByPageID(pageID int64) ([]domain.Block, error)
	GetByID(id int64) (domain.Block, error)
	Create(pageID int64, title, blockType string, content, readmore, image, imageHover *string) (int64, error)
	Update(id int64, title, content, readmore, image, imageHover *string) error
	UpdateSort(pageID int64, sort []int) error
	Delete(id int64) error
}

type FileBlock interface {
	GetByID(id int64) (domain.FileBlock, error)
	GetByPageID(pageID int64) (domain.FileBlock, error)
	Create(pageID int64, path string) (int64, error)
	Update(id int64, path string) error
}

type File interface {
	GetFiles(path string) ([]fs.FileInfo, error)
	DeleteFiles(path string) error
	CreateFolder(name string, path string) error
	RenameFile(fullPath string, newFullPath string) error
}

type Address interface {
	GetByID(id int64) (domain.Address, error)
	GetByOrganizationID(organizationID int64) ([]domain.Address, error)
	Create(organizationID int64, firstname, lastname, email, phone string, phoneExtra, role, info null.String) (int64, error)
	Update(id int64, firstname, lastname, email, phone, phoneExtra, role, info null.String) error
	UpdateSort(organizationID int64, sort []int64) error
	Delete(id int64) error
}

type Email interface {
	SendAdminNewOrganization(organizationID int64, plan, email, name, organizationRole, organizationName, website, city string, population int, phone, address, invoiceAddress, notes string) error
	SendUserNewOrganization(organizationID int64, plan, email, name, organizationRole, organizationName, website, city string, population int, phone, address, invoiceAddress, notes string) error
	SendUserVerificatonLink(emailTo, token string) error
	SendUserVerificatonLinkWithInvite(emailTo, token string, organizationName string) error
	SendUserInvite(emailTo string, organizationID int64, organizationName string) error
	SendPasswordResetLink(emailTo, token string) error
}

type Service struct {
	Email
	Auth
	User
	Organization
	OrganizationUser
	Page
	Block
	FileBlock
	File
	Address
}

func NewService(repo *repository.Repository, cache *cache.MemoryCache) *Service {
	emailService := NewEmailService()
	organizationService := NewOrganizationService(repo.Organization)
	return &Service{
		Email:            emailService,
		Auth:             NewAuthService(repo.User, repo.OrganizationUser),
		User:             NewUserService(repo.User, organizationService, emailService),
		Organization:     organizationService,
		OrganizationUser: NewOrganizationUserService(repo.OrganizationUser),
		Page:             NewPageService(repo.Page),
		Block:            NewBlockService(repo.Block),
		FileBlock:        NewFileBlockService(repo.FileBlock),
		File:             NewFileService(),
		Address:          NewAddressService(repo.Address),
	}
}
