package domain

type OrganizationUser struct {
	OrganizationID int64  `db:"organization_id" json:"organizationID"`
	UserID         int64  `db:"user_id" json:"userID"`
	Role           string `db:"role" json:"role"`
}
