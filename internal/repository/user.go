package repository

import (
	"database/sql"
	"strings"
	"time"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

type UserRepo struct {
	db               *sqlx.DB
	organizationRepo *OrganizationRepo
}

func NewUserRepo(db *sqlx.DB, organizationRepo *OrganizationRepo) *UserRepo {
	return &UserRepo{db: db, organizationRepo: organizationRepo}
}

func (r *UserRepo) GetAll() ([]domain.User, error) {
	var users []domain.User
	err := r.db.Select(&users, "SELECT * FROM users")
	return users, err
}

func (r *UserRepo) GetByID(id int64) (domain.User, error) {
	var user domain.User
	err := r.db.Get(&user, "SELECT * FROM users WHERE id = ?", id)
	return user, err
}

func (r *UserRepo) GetByEmail(email string) (domain.User, error) {
	var user domain.User
	err := r.db.Get(&user, "SELECT * FROM users WHERE email = ?", email)
	return user, err
}

func (r *UserRepo) GetByVerificationToken(token string) (domain.User, error) {
	var user domain.User
	err := r.db.Get(&user, `SELECT * FROM users WHERE id = (
			SELECT id FROM user_verifications WHERE token = ?
		)`, token)
	return user, err
}

func (r *UserRepo) GetByPasswordResetToken(token string) (domain.User, error) {
	var user domain.User
	err := r.db.Get(&user, `SELECT * FROM users WHERE id = (
			SELECT id FROM password_reset WHERE token = ?
		)`, token)
	return user, err
}

func (r *UserRepo) GetByOrganizationID(organizationID int64) ([]domain.User, error) {
	var users []domain.User
	err := r.db.Select(&users, `
		SELECT
			users.*
		FROM
			users
		INNER JOIN organizations_users AS ou
		ON
			ou.user_id = users.id AND ou.organization_id = ?
		`, organizationID)
	return users, err
}

func (r *UserRepo) GetVerification(id int64) (domain.UserVerification, error) {
	var verification domain.UserVerification
	err := r.db.Get(&verification, "SELECT * FROM user_verifications WHERE id = ?", id)
	return verification, err
}

func (r *UserRepo) GetPasswordReset(email string) (domain.UserPasswordReset, error) {
	var reset domain.UserPasswordReset
	err := r.db.Get(&reset, "SELECT * FROM password_reset WHERE email = ?", email)
	return reset, err
}

func (r *UserRepo) GetRefreshSession(refreshToken string) (domain.RefreshSession, error) {
	var session domain.RefreshSession
	err := r.db.Get(&session, "SELECT * FROM refresh_sessions WHERE refresh_token = ?", refreshToken)
	return session, err
}

func (r *UserRepo) Create(firstname, lastname *string, email string, image *string) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	userID, err := r.CreateTx(tx, firstname, lastname, email, image)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (r *UserRepo) CreateWithOrganization(firstname, lastname *string, email string, name, city, address, invoiceAddress, plan string, population int) (userID int64, organizationID int64, err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, 0, err
	}

	userID, err = r.CreateTx(tx, firstname, lastname, email, nil)
	if err != nil {
		tx.Rollback()
		return 0, 0, err
	}

	organizationID, err = r.organizationRepo.CreateTx(tx, nil, name, city, address, invoiceAddress, plan, population, userID)
	if err != nil {
		tx.Rollback()
		return 0, 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, 0, err
	}

	go r.organizationRepo.CopyDefaultContent(organizationID, plan)

	return userID, organizationID, nil
}

func (r *UserRepo) CreateTx(tx *sql.Tx, firstname, lastname *string, email string, image *string) (int64, error) {
	result, err := tx.Exec(`INSERT INTO users (firstname, lastname, email, image) VALUES (?, ?, ?, ?)`, firstname, lastname, email, image)
	if err != nil {
		return 0, err
	}

	userID, _ := result.LastInsertId()

	return userID, nil
}

func (r *UserRepo) CreateVerification(id int64, token string) error {
	_, err := r.db.Exec(`INSERT INTO user_verifications (id, token, status) VALUES (?, ?, "not_verified")`, id, token)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepo) CreatePasswordReset(email string, token string) error {
	_, err := r.db.Exec(`
		INSERT INTO password_reset (id, token) 
		VALUES (
			(SELECT id from users WHERE email = ?)
			, ?
		) ON DUPLICATE KEY UPDATE token = ?`, email, token, token)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepo) CreateRefreshSession(userID int64, refreshToken, ua, fingerprint string, expiresAt time.Time) error {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM refresh_sessions WHERE user_id = ?`, userID)
	if err != nil {
		return err
	}

	// If session count more than 7, then delete the oldest session.
	if count >= 7 {
		_, err := r.db.Exec(`DELETE FROM refresh_sessions WHERE user_id = ? ORDER BY created_at ASC LIMIT 1`, userID)
		if err != nil {
			return err
		}
	}

	_, err = r.db.Exec(`INSERT INTO refresh_sessions (user_id, refresh_token, ua, fingerprint, expires_at) VALUES (?, ?, ?, ?, ?)`, userID, refreshToken, ua, fingerprint, expiresAt)
	return err
}

func (r *UserRepo) CreateOrganizationUser(organizationID, userID int64, role string) error {
	_, err := r.db.Exec(`
		INSERT INTO organizations_users (organization_id, user_id, role) 
		VALUES (?, ?, ?)
		`, organizationID, userID, role)

	return err
}

func (r *UserRepo) Update(id int64, image, firstname, lastname, email, userType *string) error {
	updateFields := make([]string, 0)
	updateArgs := make([]interface{}, 0)

	if image != nil {
		updateFields = append(updateFields, "image = ?")
		updateArgs = append(updateArgs, *image)
	}
	if firstname != nil {
		updateFields = append(updateFields, "firstname = ?")
		updateArgs = append(updateArgs, *firstname)
	}
	if lastname != nil {
		updateFields = append(updateFields, "lastname = ?")
		updateArgs = append(updateArgs, *lastname)
	}
	if email != nil {
		updateFields = append(updateFields, "email = ?")
		updateArgs = append(updateArgs, *email)
	}
	if userType != nil {
		updateFields = append(updateFields, "type = ?")
		updateArgs = append(updateArgs, *userType)
	}

	if len(updateFields) == 0 {
		return nil
	}

	fields := strings.Join(updateFields, ",")
	updateArgs = append(updateArgs, id)

	_, err := r.db.Exec(`UPDATE users SET `+fields+` WHERE id = ?`, updateArgs...)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepo) UpdateOrganizationUser(organizationID, userID int64, role string) error {
	_, err := r.db.Exec(`UPDATE organizations_users SET role = ? WHERE organization_id = ? AND user_id = ?`, role, organizationID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepo) UpdatePassword(id int64, password string) error {
	_, err := r.db.Exec(`UPDATE users SET password = ? WHERE id = ?`, password, id)
	return err
}

func (r *UserRepo) UpdateVerificationByToken(token, status string) error {
	_, err := r.db.Exec(`UPDATE user_verifications SET status = ? WHERE token = ?`, status, token)
	return err
}

func (r *UserRepo) UpdateRefreshSession(id int64, refreshToken, ua, fingerprint string, expiresAt time.Time) error {
	_, err := r.db.Exec(`UPDATE refresh_sessions SET refresh_token = ?, ua = ?, fingerprint = ?, expires_at = ?, created_at = ? WHERE id = ?`, refreshToken, ua, fingerprint, expiresAt, time.Now(), id)
	return err
}

func (r *UserRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	return err
}

func (r *UserRepo) DeleteOrganizationUser(organizationID, userID int64) error {
	_, err := r.db.Exec(`DELETE FROM organizations_users WHERE organization_id = ? AND user_id = ?`, organizationID, userID)
	return err
}
