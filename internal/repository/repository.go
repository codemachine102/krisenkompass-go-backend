package repository

import (
	"database/sql"
	"time"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"
)

type User interface {
	GetAll() ([]domain.User, error)
	GetByID(id int64) (domain.User, error)
	GetByEmail(email string) (domain.User, error)
	GetByVerificationToken(token string) (domain.User, error)
	GetByPasswordResetToken(token string) (domain.User, error)
	GetByOrganizationID(organizationID int64) ([]domain.User, error)
	GetVerification(id int64) (domain.UserVerification, error)
	GetPasswordReset(email string) (domain.UserPasswordReset, error)
	GetRefreshSession(refreshToken string) (domain.RefreshSession, error)
	Create(firstname, lastname *string, email string, image *string) (int64, error)
	CreateWithOrganization(firstname, lastname *string, email string, name, city, address, invoiceAddress, plan string, population int) (userID int64, organizationID int64, err error)
	CreateTx(tx *sql.Tx, firstname, lastname *string, email string, image *string) (int64, error)
	CreateVerification(id int64, token string) error
	CreatePasswordReset(email string, token string) error
	CreateRefreshSession(userID int64, refreshToken, ua, fingerprint string, expiresAt time.Time) error
	CreateOrganizationUser(organizationID, userID int64, role string) error
	Update(id int64, image, firstname, lastname, email, userType *string) error
	UpdatePassword(id int64, password string) error
	UpdateVerificationByToken(token, status string) error
	UpdateRefreshSession(id int64, refreshToken, ua, fingerprint string, expiresAt time.Time) error
	UpdateOrganizationUser(organizationID, userID int64, role string) error
	Delete(id int64) error
	DeleteOrganizationUser(organizationID, userID int64) error
}

type Organization interface {
	GetAll() ([]domain.Organization, error)
	GetByID(id int64) (domain.Organization, error)
	GetByUserID(userID int64) ([]domain.Organization, error)
	Create(image *string, name, city, address, invoiceAddress, plan string, population int, userID int64) (int64, error)
	CreateTx(tx *sql.Tx, image *string, name, city, address, invoiceAddress, plan string, population int, userID int64) (int64, error)
	CopyDefaultContent(organizationID int64, plan string) error
	Update(id int64, image, name, city, address, invoiceAddress, plan, status null.String, population null.Int) error
	Delete(id int64) error
}

type OrganizationUser interface {
	GetByUserID(id int64) ([]domain.OrganizationUser, error)
}

type Page interface {
	GetByID(id int64, fields ...string) (domain.Page, error)
	GetChildrens(id int64) ([]domain.Page, error)
	GetDefaultPage(pageID int64) ([]domain.DefaultPage, error)
	GetRootPages(organizationID *int64) ([]domain.Page, error)
	GetPages(organizationID *int64, fields ...string) ([]domain.Page, error)
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

type Address interface {
	GetByID(id int64) (domain.Address, error)
	GetByOrganizationID(organizationID int64) ([]domain.Address, error)
	Create(organizationID int64, firstname, lastname, email, phone string, phoneExtra, role, info null.String) (int64, error)
	Update(id int64, firstname, lastname, email, phone, phoneExtra, role, info null.String) error
	UpdateSort(organizationID int64, sort []int64) error
	Delete(id int64) error
}

type Repository struct {
	User
	Organization
	OrganizationUser
	Page
	Block
	FileBlock
	Address
}

func NewRepository(db *sqlx.DB) *Repository {
	pageRepo := NewPageRepo(db)
	blockRepo := NewBlockRepo(db, pageRepo)
	fileBlockRepo := NewFileRepo(db, pageRepo)
	organizationRepo := NewOrganizationRepo(db, pageRepo, blockRepo, fileBlockRepo)
	return &Repository{
		User:             NewUserRepo(db, organizationRepo),
		OrganizationUser: NewOrganizationUserRepo(db),
		Page:             pageRepo,
		Block:            blockRepo,
		Organization:     organizationRepo,
		FileBlock:        fileBlockRepo,
		Address:          NewAddressRepo(db),
	}
}
