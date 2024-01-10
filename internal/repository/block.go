package repository

import (
	"errors"
	"strings"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/domain"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"
)

type BlockRepo struct {
	db       *sqlx.DB
	pageRepo *PageRepo
}

func NewBlockRepo(db *sqlx.DB, pageRepo *PageRepo) *BlockRepo {
	return &BlockRepo{db: db, pageRepo: pageRepo}
}

func (r *BlockRepo) GetByID(id int64) (domain.Block, error) {
	var block domain.Block
	err := r.db.Get(&block, "SELECT * FROM blocks WHERE id = ?", id)
	return block, err
}

func (r *BlockRepo) GetByPageID(pageID int64) ([]domain.Block, error) {
	var blocks []domain.Block
	err := r.db.Select(&blocks, "SELECT * FROM blocks WHERE page_id = ? ORDER BY sort", pageID)
	return blocks, err
}

func (r *BlockRepo) Create(pageID int64, title, blockType string, content, readmore, image, imageHover *string) (int64, error) {
	cnt := null.StringFromPtr(content)
	rdmr := null.StringFromPtr(readmore)
	img := null.StringFromPtr(image)
	imgHover := null.StringFromPtr(imageHover)

	result, err := r.db.Exec(`INSERT 
		INTO blocks (page_id, title, content, readmore, image, image_hover, type, sort) 
		VALUES (?, ?, ?, ?, ?, ?, ?, (
			SELECT COALESCE((MAX(B.sort) + 1), 1) 
			FROM blocks AS B 
			WHERE B.page_id = ?)
		)`,
		pageID, title, cnt, rdmr, img, imgHover, blockType, pageID)
	if err != nil {
		return 0, err
	}
	blockID, _ := result.LastInsertId()

	go r.pageRepo.UpdateUpdatedAt(pageID)

	return blockID, nil
}

func (r *BlockRepo) Update(id int64, title, content, readmore, image, imageHover *string) error {
	fields := make([]string, 0)
	args := make([]interface{}, 0)

	if content == nil && title == nil && image == nil && imageHover == nil {
		return errors.New("no input data specified")
	}

	if content != nil {
		fields = append(fields, "content = ?")
		args = append(args, *content)
	}

	if readmore != nil {
		fields = append(fields, "readmore = ?")
		args = append(args, *readmore)
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
	query := "UPDATE blocks SET " + strings.Join(fields, ", ") + " WHERE id = ?"

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return err
	}

	go func() {
		block, err := r.GetByID(id)
		if err == nil {
			r.pageRepo.UpdateUpdatedAt(block.PageID)
		}
	}()

	return nil
}

func (r *BlockRepo) UpdateSort(pageID int64, sort []int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	for i := range sort {
		_, err = tx.Exec("UPDATE blocks SET sort = ? WHERE id = ? AND page_id = ?", i, sort[i], pageID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	go r.pageRepo.UpdateUpdatedAt(pageID)

	return tx.Commit()
}

func (r *BlockRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM blocks WHERE id = ?", id)
	return err
}
