package repository

import (
	"strings"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"
)

type AddressRepo struct {
	db *sqlx.DB
}

func NewAddressRepo(db *sqlx.DB) *AddressRepo {
	return &AddressRepo{db: db}
}

func (r *AddressRepo) GetByID(id int64) (domain.Address, error) {
	var address domain.Address
	err := r.db.Get(&address, "SELECT * FROM addresses WHERE id = ?", id)
	return address, err
}

func (r *AddressRepo) GetByOrganizationID(organizationID int64) ([]domain.Address, error) {
	var addresses []domain.Address
	err := r.db.Select(&addresses, "SELECT * FROM addresses WHERE organization_id = ? ORDER BY sort", organizationID)
	return addresses, err
}

func (r *AddressRepo) Create(organizationID int64, firstname, lastname, email, phone string, phoneExtra, role, info null.String) (int64, error) {
	result, err := r.db.Exec(`INSERT INTO addresses (organization_id, firstname, lastname, email, phone, phone_extra, role, info, sort) VALUES (?, ?, ?, ?, ?, ?, ?, ?, (
			SELECT COALESCE((MAX(A.sort) + 1), 1) 
			FROM addresses AS A
			WHERE A.organization_id = ?)
		)`, organizationID, firstname, lastname, email, phone, phoneExtra, role, info, organizationID)
	if err != nil {
		return 0, err
	}

	id, _ := result.LastInsertId()
	return id, nil
}

func (r *AddressRepo) Update(id int64, firstname, lastname, email, phone, phoneExtra, role, info null.String) error {
	updateFields := make([]string, 0)
	updateArgs := make([]interface{}, 0)

	if firstname.Valid {
		updateFields = append(updateFields, "firstname = ?")
		updateArgs = append(updateArgs, firstname.String)
	}
	if lastname.Valid {
		updateFields = append(updateFields, "lastname = ?")
		updateArgs = append(updateArgs, lastname.String)
	}
	if email.Valid {
		updateFields = append(updateFields, "email = ?")
		updateArgs = append(updateArgs, email.String)
	}
	if phone.Valid {
		updateFields = append(updateFields, "phone = ?")
		updateArgs = append(updateArgs, phone.String)
	}
	if phoneExtra.Valid {
		updateFields = append(updateFields, "phone_extra = ?")
		updateArgs = append(updateArgs, phoneExtra.String)
	}
	if role.Valid {
		updateFields = append(updateFields, "role = ?")
		updateArgs = append(updateArgs, role.String)
	}
	if info.Valid {
		updateFields = append(updateFields, "info = ?")
		updateArgs = append(updateArgs, info.String)
	}

	if len(updateFields) == 0 {
		return nil
	}

	fields := strings.Join(updateFields, ",")
	updateArgs = append(updateArgs, id)

	_, err := r.db.Exec(`UPDATE addresses SET `+fields+` WHERE id = ?`, updateArgs...)
	if err != nil {
		return err
	}

	return nil
}

func (r *AddressRepo) UpdateSort(organizationID int64, sort []int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	for i := range sort {
		_, err = tx.Exec("UPDATE addresses SET sort = ? WHERE id = ? AND organization_id = ?", i, sort[i], organizationID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *AddressRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM addresses WHERE id = ?", id)
	return err
}
