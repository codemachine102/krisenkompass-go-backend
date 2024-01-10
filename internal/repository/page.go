package repository

import (
	"fmt"
	"strings"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"
)

type PageRepo struct {
	db *sqlx.DB
}

func NewPageRepo(db *sqlx.DB) *PageRepo {
	return &PageRepo{db: db}
}

func (r *PageRepo) GetByID(id int64, fields ...string) (domain.Page, error) {
	var page domain.Page
	err := r.db.Get(&page, "SELECT * FROM pages WHERE id = ?", id)
	return page, err
}

func (r *PageRepo) GetChildrens(id int64) ([]domain.Page, error) {
	var pages []domain.Page
	err := r.db.Select(&pages, "SELECT * FROM pages WHERE parent_id = ? ORDER BY sort", id)
	return pages, err
}

func (r *PageRepo) GetDefaultPage(pageID int64) ([]domain.DefaultPage, error) {
	var pages []domain.DefaultPage
	err := r.db.Select(&pages, "SELECT * FROM default_pages WHERE id = ?", pageID)
	return pages, err
}

func (r *PageRepo) GetRootPages(organizationID *int64) ([]domain.Page, error) {
	var pages []domain.Page
	var err error
	if organizationID == nil {
		err = r.db.Select(&pages, "SELECT * FROM pages WHERE organization_id IS NULL AND parent_id IS NULL")
	} else {
		err = r.db.Select(&pages, "SELECT * FROM pages WHERE organization_id = ? AND parent_id IS NULL", organizationID)
	}
	return pages, err
}

func (r *PageRepo) GetPages(organizationID *int64, fields ...string) ([]domain.Page, error) {
	var pages []domain.Page
	var selectFields string
	var err error

	if len(fields) == 0 {
		selectFields = "*"
	}

	for _, field := range fields {
		selectFields += fmt.Sprintf("pages.%s,", field)
	}

	selectFields = strings.Trim(selectFields, ", ")

	if organizationID == nil {
		err = r.db.Select(&pages, "SELECT "+selectFields+" FROM pages WHERE organization_id IS NULL ORDER BY pages.sort")
	} else {
		err = r.db.Select(&pages, "SELECT "+selectFields+" FROM pages WHERE organization_id = ? ORDER BY pages.sort", organizationID)
	}
	return pages, err
}

func (r *PageRepo) Create(organizationID *int64, parentID int64, languageTag, pageType, theme, status, title string, image, imageHover *string) (int64, error) {
	NullOrganizationID := null.IntFromPtr(organizationID)
	img := null.StringFromPtr(image)
	imgHover := null.StringFromPtr(imageHover)

	result, err := r.db.Exec(`INSERT 
			INTO pages (organization_id, parent_id, language_tag, pages.type, theme, pages.status, title, image, image_hover, pages.sort)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, (
				SELECT COALESCE((MAX(P.sort) + 1), 1) 
				FROM pages AS P
				WHERE P.parent_id = ?))`,
		NullOrganizationID, parentID, languageTag, pageType, theme, status, title, img, imgHover, parentID)

	if err != nil {
		return 0, err
	}
	pageID, _ := result.LastInsertId()

	go r.UpdateUpdatedAt(parentID)

	return pageID, nil
}

func (r *PageRepo) CreateDefaultPage(pageID int64, plans []string) error {
	if len(plans) == 0 {
		return nil
	}

	var insertValues string
	var insertArgs = make([]interface{}, 0)

	for _, p := range plans {
		insertValues += "(?, ?),"
		insertArgs = append(insertArgs, pageID, p)
	}

	insertValues = strings.TrimSuffix(insertValues, ",")

	_, err := r.db.Exec(`INSERT INTO default_pages (id, plan) VALUES `+insertValues+` ON DUPLICATE KEY UPDATE id=id`, insertArgs...)

	return err
}

func (r *PageRepo) Update(id int64, parentID *int64, status, title, image, imageHover *string) error {
	fields := make([]string, 0)
	args := make([]interface{}, 0)

	if status == nil && title == nil && image == nil && imageHover == nil && parentID == nil {
		return nil
	}

	if parentID != nil {
		fields = append(fields, "parent_id = ?")
		args = append(args, *parentID)

		var theme string
		err := r.db.Get(&theme, "SELECT pages.theme FROM pages WHERE id = ?", parentID)
		if err != nil {
			return err
		}
	}

	if status != nil {
		fields = append(fields, "status = ?")
		args = append(args, *status)
	}

	if title != nil {
		fields = append(fields, "title = ?")
		args = append(args, *title)
	}

	if image != nil {
		fields = append(fields, "image = ?")
		args = append(args, *image)
	}

	if imageHover != nil {
		fields = append(fields, "image_hover = ?")
		args = append(args, *imageHover)
	}

	args = append(args, id)
	query := "UPDATE pages SET updated_at = NOW(), " + strings.Join(fields, ", ") + " WHERE id = ?"

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *PageRepo) UpdateSort(parentID int64, sort []int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	for i := range sort {
		_, err = tx.Exec("UPDATE pages SET sort = ? WHERE id = ? AND parent_id = ?", i, sort[i], parentID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	go r.UpdateUpdatedAt(parentID)

	return tx.Commit()
}

func (r *PageRepo) UpdateUpdatedAt(pageID int64) error {
	_, err := r.db.Exec("UPDATE pages SET updated_at = NOW() WHERE id = ?", pageID)

	return err
}

func (r *PageRepo) Delete(id int64) error {
	page, err := r.GetByID(id)
	if err == nil {
		go r.UpdateUpdatedAt(page.ParentID.Int64)
	}

	_, err = r.db.Exec("DELETE FROM pages WHERE id = ?", id)

	return err
}

func (r *PageRepo) DeleteDefaultPage(pageID int64, plans []string) error {
	if len(plans) == 0 {
		return nil
	}

	planValues := "\"" + strings.Join(plans, "\",\"") + "\""

	_, err := r.db.Exec("DELETE FROM default_pages WHERE id = ? AND plan IN ("+planValues+")", pageID)
	return err
}
