package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sprucehealth/backend/common"
)

func (d *DataService) GetResourceGuide(id int64) (*common.ResourceGuide, error) {
	var guide common.ResourceGuide
	var layout []byte
	row := d.db.QueryRow(`SELECT id, section_id, ordinal, title, photo_url, layout FROM resource_guide WHERE id = ?`, id)
	err := row.Scan(
		&guide.ID,
		&guide.SectionID,
		&guide.Ordinal,
		&guide.Title,
		&guide.PhotoURL,
		&layout,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound("resource_guide")
	} else if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(layout, &guide.Layout); err != nil {
		return nil, err
	}
	return &guide, nil
}

func (d *DataService) ListResourceGuideSections() ([]*common.ResourceGuideSection, error) {
	rows, err := d.db.Query(`SELECT id, ordinal, title FROM resource_guide_section ORDER BY ordinal`)
	if err != nil {
		return nil, err
	}
	defer rows.Next()
	var sections []*common.ResourceGuideSection
	for rows.Next() {
		var sec common.ResourceGuideSection
		err := rows.Scan(
			&sec.ID,
			&sec.Ordinal,
			&sec.Title,
		)
		if err != nil {
			return nil, err
		}
		sections = append(sections, &sec)
	}
	return sections, rows.Err()
}

func (d *DataService) ListResourceGuides(withLayouts bool) ([]*common.ResourceGuideSection, map[int64][]*common.ResourceGuide, error) {
	sections, err := d.ListResourceGuideSections()
	if err != nil {
		return nil, nil, err
	}

	layoutCol := ""
	if withLayouts {
		layoutCol = ", layout"
	}

	rows, err := d.db.Query(`SELECT id, section_id, ordinal, title, photo_url` + layoutCol + ` FROM resource_guide ORDER BY ordinal`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Next()
	guides := map[int64][]*common.ResourceGuide{}
	var values []interface{}
	for rows.Next() {
		var guide common.ResourceGuide
		var layout []byte
		values = append(values[:0],
			&guide.ID,
			&guide.SectionID,
			&guide.Ordinal,
			&guide.Title,
			&guide.PhotoURL)
		if withLayouts {
			values = append(values, &layout)
		}
		if err := rows.Scan(values...); err != nil {
			return nil, nil, err
		}
		if layout != nil {
			if err := json.Unmarshal(layout, &guide.Layout); err != nil {
				return nil, nil, err
			}
		}
		guides[guide.SectionID] = append(guides[guide.SectionID], &guide)
	}
	return sections, guides, rows.Err()
}

func (d *DataService) ReplaceResourceGuides(sections []*common.ResourceGuideSection, guides map[int64][]*common.ResourceGuide) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	err = func(tx *sql.Tx) error {
		if _, err := tx.Exec(`DELETE FROM resource_guide`); err != nil {
			return err
		}
		if _, err := tx.Exec(`DELETE FROM resource_guide_section`); err != nil {
			return err
		}
		insertSection, err := tx.Prepare(`INSERT INTO resource_guide_section (id, title, ordinal) VALUEs (?, ?, ?)`)
		if err != nil {
			return err
		}
		defer insertSection.Close()
		insertGuide, err := tx.Prepare(`INSERT INTO resource_guide (id, title, section_id, ordinal, photo_url, layout) VALUEs (?, ?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		defer insertGuide.Close()
		for _, s := range sections {
			if _, err := insertSection.Exec(s.ID, s.Title, s.Ordinal); err != nil {
				return err
			}
		}
		for secID, gs := range guides {
			for _, g := range gs {
				layout, err := json.Marshal(g.Layout)
				if err != nil {
					return err
				}
				if _, err := insertGuide.Exec(g.ID, g.Title, secID, g.Ordinal, g.PhotoURL, layout); err != nil {
					return err
				}
			}
		}
		return nil
	}(tx)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *DataService) CreateResourceGuideSection(sec *common.ResourceGuideSection) (int64, error) {
	if sec.Title == "" || sec.Ordinal == 0 {
		return 0, fmt.Errorf("api.CreateResourceGuideSection: Title and Ordinal may not be empty")
	}
	res, err := d.db.Exec("INSERT INTO resource_guide_section (title, ordinal) VALUES (?, ?)", sec.Title, sec.Ordinal)
	if err != nil {
		return 0, err
	}
	sec.ID, err = res.LastInsertId()
	return sec.ID, err
}

func (d *DataService) UpdateResourceGuideSection(sec *common.ResourceGuideSection) error {
	if sec.ID <= 0 {
		return fmt.Errorf("api.UpdateResourceGuideSection: ID may not be 0")
	}
	var columns []string
	var values []interface{}
	if sec.Title != "" {
		columns = append(columns, "title = ?")
		values = append(values, sec.Title)
	}
	if sec.Ordinal > 0 {
		columns = append(columns, "ordinal = ?")
		values = append(values, sec.Ordinal)
	}
	if len(columns) == 0 {
		return fmt.Errorf("api.UpdateResourceGuideSection: nothing to update")
	}
	values = append(values, sec.ID)
	_, err := d.db.Exec("UPDATE resource_guide_section SET "+strings.Join(columns, ",")+" WHERE id = ?", values...)
	return err
}

func (d *DataService) CreateResourceGuide(guide *common.ResourceGuide) (int64, error) {
	if guide.Title == "" || guide.PhotoURL == "" || guide.Layout == nil {
		return 0, fmt.Errorf("api.CreateResourceGuide: Title, PhotoURL, and Layout may not be empty")
	}
	layout, err := json.Marshal(guide.Layout)
	if err != nil {
		return 0, err
	}
	res, err := d.db.Exec("INSERT INTO resource_guide (title, section_id, ordinal, photo_url, layout) VALUES (?, ?, ?, ?, ?)",
		guide.Title, guide.SectionID, guide.Ordinal, guide.PhotoURL, layout)
	if err != nil {
		return 0, err
	}
	guide.ID, err = res.LastInsertId()
	return guide.ID, err
}

func (d *DataService) UpdateResourceGuide(guide *common.ResourceGuide) error {
	if guide.ID <= 0 {
		return fmt.Errorf("api.UpdateResourceGuide: ID may not be 0")
	}
	var columns []string
	var values []interface{}
	if guide.Title != "" {
		columns = append(columns, "title = ?")
		values = append(values, guide.Title)
	}
	if guide.SectionID != 0 {
		columns = append(columns, "section_id = ?")
		values = append(values, guide.SectionID)
	}
	if guide.Ordinal > 0 {
		columns = append(columns, "ordinal = ?")
		values = append(values, guide.Ordinal)
	}
	if guide.PhotoURL != "" {
		columns = append(columns, "photo_url = ?")
		values = append(values, guide.PhotoURL)
	}
	if guide.Layout != nil {
		columns = append(columns, "layout = ?")
		b, err := json.Marshal(guide.Layout)
		if err != nil {
			return err
		}
		values = append(values, b)
	}
	if len(columns) == 0 {
		return fmt.Errorf("api.UpdateResourceGuide: nothing to update")
	}
	values = append(values, guide.ID)
	_, err := d.db.Exec("UPDATE resource_guide SET "+strings.Join(columns, ",")+" WHERE id = ?", values...)
	return err
}
