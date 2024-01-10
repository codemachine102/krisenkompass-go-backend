package repository

import (
	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

type OrganizationUserRepo struct {
	db *sqlx.DB
}

func NewOrganizationUserRepo(db *sqlx.DB) *OrganizationUserRepo {
	return &OrganizationUserRepo{db: db}
}

func (r *OrganizationUserRepo) GetByUserID(id int64) ([]domain.OrganizationUser, error) {
	var rows []domain.OrganizationUser
	err := r.db.Select(&rows, "SELECT * FROM organizations_users WHERE user_id = ?", id)
	return rows, err
}
