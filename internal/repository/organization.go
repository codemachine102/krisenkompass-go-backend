package repository

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"
)

type OrganizationRepo struct {
	db            *sqlx.DB
	pageRepo      *PageRepo
	blockRepo     *BlockRepo
	fileBlockRepo *FileRepo
}

func NewOrganizationRepo(db *sqlx.DB, pageRepo *PageRepo, blockRepo *BlockRepo, fileBlockRepo *FileRepo) *OrganizationRepo {
	return &OrganizationRepo{db: db, pageRepo: pageRepo, blockRepo: blockRepo, fileBlockRepo: fileBlockRepo}
}

func (r *OrganizationRepo) GetAll() ([]domain.Organization, error) {
	var organizations []domain.Organization
	err := r.db.Select(&organizations, "SELECT * FROM organizations")
	return organizations, err
}

func (r *OrganizationRepo) GetByID(id int64) (domain.Organization, error) {
	var organization domain.Organization
	err := r.db.Get(&organization, "SELECT * FROM organizations WHERE id = ?", id)
	return organization, err
}

func (r *OrganizationRepo) GetByUserID(userID int64) ([]domain.Organization, error) {
	var organizations []domain.Organization
	err := r.db.Select(&organizations, `
		SELECT
			organizations.*
		FROM
			organizations
		INNER JOIN organizations_users AS ou
		ON
			ou.organization_id = organizations.id AND ou.user_id = ?
		`, userID)
	return organizations, err
}

func (r *OrganizationRepo) Create(image *string, name, city, address, invoiceAddress, plan string, population int, userID int64) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	organizationID, err := r.CreateTx(tx, image, name, city, address, invoiceAddress, plan, population, userID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	go r.CopyDefaultContent(organizationID, plan)

	return organizationID, nil
}

func (r *OrganizationRepo) CreateTx(tx *sql.Tx, image *string, name, city, address, invoiceAddress, plan string, population int, userID int64) (int64, error) {
	result, err := tx.Exec(`INSERT INTO organizations (name, image, city, population, address, invoice_address, plan) VALUES (?, ?, ?, ?, ?, ?, ?)`, name, image, city, population, address, invoiceAddress, plan)
	if err != nil {
		return 0, err
	}

	organizationID, _ := result.LastInsertId()

	_, err = tx.Exec(`
		INSERT INTO organizations_users (organization_id, user_id, role) 
		VALUES (?, ?, "owner")
		`, organizationID, userID)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(`INSERT 
		INTO pages (organization_id, language_tag, pages.type, theme, pages.status, title, sort) 
		VALUES (?, ?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?, ?)`,
		organizationID, "de", "section", "precautions", "visible", "Vorsorgen", 1,
		organizationID, "de", "section", "deal_with", "visible", "Bewältigen", 2,
		organizationID, "de", "section", "e_restore", "visible", "Wiederherstellen", 3,
		organizationID, "de", "section", "e_avoid", "visible", "Vermeiden", 4,
		organizationID, "de", "section", "e_gfs", "visible", "GFS", 5,
		organizationID, "de", "section", "e_school", "visible", "Schulen", 6,
	)
	if err != nil {
		return 0, err
	}

	return organizationID, nil
}

func (r *OrganizationRepo) Update(id int64, name, image, city, address, invoiceAddress, plan, status null.String, population null.Int) error {
	updateFields := make([]string, 0)
	updateArgs := make([]interface{}, 0)

	if image.Valid {
		updateFields = append(updateFields, "image = ?")
		updateArgs = append(updateArgs, image.String)
	}
	if name.Valid {
		updateFields = append(updateFields, "name = ?")
		updateArgs = append(updateArgs, name.String)
	}
	if city.Valid {
		updateFields = append(updateFields, "city = ?")
		updateArgs = append(updateArgs, city.String)
	}
	if address.Valid {
		updateFields = append(updateFields, "address = ?")
		updateArgs = append(updateArgs, address.String)
	}
	if invoiceAddress.Valid {
		updateFields = append(updateFields, "invoice_address = ?")
		updateArgs = append(updateArgs, invoiceAddress.String)
	}
	if plan.Valid {
		updateFields = append(updateFields, "plan = ?")
		updateArgs = append(updateArgs, plan.String)
	}
	if status.Valid {
		updateFields = append(updateFields, "status = ?")
		updateArgs = append(updateArgs, status.String)
	}
	if population.Valid {
		updateFields = append(updateFields, "population = ?")
		updateArgs = append(updateArgs, population.Int64)
	}

	if len(updateFields) == 0 {
		return nil
	}

	fields := strings.Join(updateFields, ",")
	updateArgs = append(updateArgs, id)

	_, err := r.db.Exec(`UPDATE organizations SET `+fields+` WHERE id = ?`, updateArgs...)
	if err != nil {
		return err
	}

	return nil
}

func (r *OrganizationRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM organizations WHERE id = ?`, id)
	return err
}

func (r *OrganizationRepo) CopyDefaultContent(organizationID int64, plan string) error {
	adminRootPages, err := r.pageRepo.GetRootPages(nil)
	if err != nil {
		return err
	}

	// Create maps to save new page and block IDs, so then page links can be fixed.
	var pageIDs sync.Map
	var blockIDs sync.Map

	c := make(chan error)
	for _, root := range adminRootPages {
		var pages []domain.Page

		err = r.db.Select(&pages, `
		SELECT pages.* FROM pages INNER JOIN default_pages ON default_pages.id = pages.id
		WHERE pages.status = "visible" 
			AND default_pages.plan = ?
			AND pages.parent_id = ? 
			AND pages.theme = ?`, plan, root.ID, root.Theme)
		if err != nil {
			log.Println("Failed to get default pages: ", err)
		}

		var newParentID int64
		err = r.db.Get(&newParentID, "SELECT id FROM pages WHERE organization_id = ? AND parent_id IS NULL AND theme = ?", organizationID, root.Theme)
		if err != nil {
			log.Println("Failed to get organization root page: ", err)
		}

		go func() {
			c <- r.copyPages(pages, newParentID, organizationID, &pageIDs, &blockIDs)
		}()
	}

	for range adminRootPages {
		err := <-c
		if err != nil {
			log.Println("Error during content copying: ", err)
		}
	}

	err = r.fixPageLinks(organizationID, &pageIDs, &blockIDs)
	if err != nil {
		log.Println("Failed to fix links: ", err)
	}

	return nil
}

func (r *OrganizationRepo) copyPages(pages []domain.Page, newParentID, organizationID int64, pageIDs *sync.Map, blockIDs *sync.Map) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	newPageIDs := make([]int64, len(pages))

	for i, page := range pages {
		result, err := tx.Exec(`INSERT 
			INTO pages (organization_id, parent_id, language_tag, pages.type, theme, pages.status, title, image, image_hover, pages.sort)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			organizationID, newParentID, page.LanguageTag, page.Type, page.Theme, page.Status, page.Title, page.Image, page.ImageHover, page.Sort)
		if err != nil {
			log.Println(err)
			return err
		}

		newID, _ := result.LastInsertId()
		newPageIDs[i] = newID

		pageIDs.Store(page.ID, newID)
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return err
	}

	for i, page := range pages {
		if page.Type == "section" {
			childrens, err := r.pageRepo.GetChildrens(page.ID)
			if err != nil {
				return err
			}

			err = r.copyPages(childrens, newPageIDs[i], organizationID, pageIDs, blockIDs)
			if err != nil {
				log.Println(err)
			}
		} else if page.Type == "file" {
			err = r.copyFile(page.ID, newPageIDs[i])
			if err != nil {
				log.Println(err)
				return err
			}

		} else if page.Type == "content" {
			err = r.copyBlocks(page.ID, newPageIDs[i], blockIDs)
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}

	return nil
}

func (r *OrganizationRepo) copyBlocks(targetID, destID int64, blockIDs *sync.Map) error {
	blocks, err := r.blockRepo.GetByPageID(targetID)
	if err != nil {
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	for _, block := range blocks {
		result, err := tx.Exec(`INSERT 
		INTO blocks (page_id, title, content, readmore, image, image_hover, type, sort) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			destID, block.Title, block.Content, block.Readmore, block.Image, block.ImageHover, block.Type, block.Sort)
		if err != nil {
			return err
		}

		newID, _ := result.LastInsertId()
		blockIDs.Store(block.ID, newID)
	}

	return tx.Commit()
}

func (r *OrganizationRepo) copyFile(targetID, destID int64) error {
	file, err := r.fileBlockRepo.GetByPageID(targetID)
	if err != nil {
		return err
	}

	_, err = r.fileBlockRepo.Create(destID, file.Path)
	if err != nil {
		return err
	}

	return nil
}

func (r *OrganizationRepo) fixPageLinks(organizationID int64, pageIDs *sync.Map, blockIDs *sync.Map) error {
	countBlocks := 0
	countPages := 0

	pageIDs.Range(func(key, value interface{}) bool {
		countPages += 1
		return true
	})

	blockIDs.Range(func(key, value interface{}) bool {
		countBlocks += 1
		return true
	})

	if countPages == 0 || countBlocks == 0 {
		return nil
	}

	selectIDs := ""
	blockIDs.Range(func(key, value interface{}) bool {
		selectIDs += strconv.FormatInt(value.(int64), 10) + ", "
		return true
	})
	selectIDs = strings.TrimRight(selectIDs, ", ")

	limit := 1
	offset := 0

	for {
		var block domain.Block

		err := r.db.Get(&block, "SELECT id, content, readmore  FROM blocks WHERE id IN ("+selectIDs+") LIMIT ? OFFSET ?", limit, offset)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			return err
		}

		tx, err := r.db.Begin()
		if err != nil {
			return err
		}

		re := regexp.MustCompile(`\/admin\/page\/([0-9]+)(#block-[0-9]+)?`)

		newContent := block.Content.String
		newReadmore := block.Readmore.String
		contentChanged := false
		readmoreChanged := false

		// Replace links inside content.
		matches := re.FindAllStringSubmatch(newContent, -1)
		for _, match := range matches {
			newLink := fmt.Sprintf("/organization/%d", organizationID)

			if len(match) > 1 {
				oldPageID, err := strconv.ParseInt(match[1], 10, 64)

				if _, ok := pageIDs.Load(oldPageID); err == nil && ok {
					contentChanged = true
					pageID, _ := pageIDs.Load(oldPageID)
					newLink += "/page/" + strconv.FormatInt(pageID.(int64), 10)

					if len(match) == 3 && strings.Contains(match[2], "#block-") {
						oldBlockID, err := strconv.ParseInt(strings.Split(match[2], "#block-")[1], 10, 64)
						if err == nil {
							blockID, ok := blockIDs.Load(oldBlockID)
							if ok {
								newLink += "#block-" + strconv.FormatInt(blockID.(int64), 10)
							}
						}
					}
				}
			}
			newContent = strings.Replace(newContent, match[0], newLink, -1)
		}

		// Replace links inside readmore.
		matches = re.FindAllStringSubmatch(newReadmore, -1)
		for _, match := range matches {
			newLink := fmt.Sprintf("/organization/%d", organizationID)

			if len(match) > 1 {
				oldPageID, err := strconv.ParseInt(match[1], 10, 64)

				if _, ok := pageIDs.Load(oldPageID); err == nil && ok {
					readmoreChanged = true
					pageID, _ := pageIDs.Load(oldPageID)
					newLink += "/page/" + strconv.FormatInt(pageID.(int64), 10)

					if len(match) == 3 && strings.Contains(match[2], "#block-") {
						oldBlockID, err := strconv.ParseInt(strings.Split(match[2], "#block-")[1], 10, 64)
						if err == nil {
							blockID, ok := blockIDs.Load(oldBlockID)
							if ok {
								newLink += "#block-" + strconv.FormatInt(blockID.(int64), 10)
							}
						}
					}
				}
			}

			newReadmore = strings.Replace(newReadmore, match[0], newLink, -1)
		}

		if contentChanged && readmoreChanged {
			_, err := tx.Exec("UPDATE blocks SET content = ?, readmore = ? WHERE id = ?", newContent, newReadmore, block.ID)
			if err != nil {
				return err
			}
		} else if contentChanged {
			_, err := tx.Exec("UPDATE blocks SET content = ? WHERE id = ?", newContent, block.ID)
			if err != nil {
				return err
			}
		} else if readmoreChanged {
			_, err := tx.Exec("UPDATE blocks SET readmore = ? WHERE id = ?", newReadmore, block.ID)
			if err != nil {
				return err
			}
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		offset++
	}

	return nil
}
