package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/dbutil"
)

func (d *dataService) Pathway(id int64, opts PathwayOption) (*common.Pathway, error) {
	if opts&POWithDetails != 0 {
		return scanPathway(opts,
			d.db.QueryRow(`SELECT id, tag, name, medicine_branch, status, details_json FROM clinical_pathway WHERE id = ?`, id))
	}
	return scanPathway(opts,
		d.db.QueryRow(`SELECT id, tag, name, medicine_branch, status FROM clinical_pathway WHERE id = ?`, id))
}

func (d *dataService) PathwayForTag(tag string, opts PathwayOption) (*common.Pathway, error) {
	if opts&POWithDetails != 0 {
		return scanPathway(opts,
			d.db.QueryRow(`SELECT id, tag, name, medicine_branch, status, details_json FROM clinical_pathway WHERE tag = ?`, tag))
	}
	return scanPathway(opts,
		d.db.QueryRow(`SELECT id, tag, name, medicine_branch, status FROM clinical_pathway WHERE tag = ?`, tag))
}

func (d *dataService) PathwaysForTags(tags []string, opts PathwayOption) (map[string]*common.Pathway, error) {
	var withDetailsQuery string
	if opts&POWithDetails != 0 {
		withDetailsQuery = ", details_json"
	}
	rows, err := d.db.Query(`
		SELECT id, tag, name, medicine_branch, status`+withDetailsQuery+`
		FROM clinical_pathway
		WHERE tag IN (`+dbutil.MySQLArgs(len(tags))+`)`,
		dbutil.AppendStringsToInterfaceSlice(nil, tags)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	pathways := make(map[string]*common.Pathway)
	for rows.Next() {
		p, err := scanPathway(opts, rows)
		if err != nil {
			return nil, err
		}
		pathways[p.Tag] = p
	}
	return pathways, rows.Err()
}

func (d *dataService) ListPathways(opts PathwayOption) ([]*common.Pathway, error) {
	var withDetailsQuery string
	if opts&POWithDetails != 0 {
		withDetailsQuery = ", details_json"
	}
	var rows *sql.Rows
	var err error
	if opts&POActiveOnly != 0 {
		rows, err = d.db.Query(`
			SELECT id, tag, name, medicine_branch, status`+withDetailsQuery+`
			FROM clinical_pathway
			WHERE status = ?
			ORDER BY name`, common.PathwayActive.String())
	} else {
		rows, err = d.db.Query(`
			SELECT id, tag, name, medicine_branch, status` + withDetailsQuery + `
			FROM clinical_pathway
			ORDER BY name`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pathways []*common.Pathway
	for rows.Next() {
		p, err := scanPathway(opts, rows)
		if err != nil {
			return nil, err
		}
		pathways = append(pathways, p)
	}
	return pathways, rows.Err()
}

func (d *dataService) CreatePathway(pathway *common.Pathway) error {
	if pathway.Tag == "" {
		return errors.New("pathway tag required")
	}
	if pathway.Name == "" {
		return errors.New("pathway name required")
	}
	if pathway.MedicineBranch == "" {
		return errors.New("pathway medicine branch required")
	}
	if pathway.Status == "" {
		return errors.New("pathway status required")
	}
	var detailsJS []byte
	if pathway.Details != nil {
		var err error
		detailsJS, err = json.Marshal(pathway.Details)
		if err != nil {
			return err
		}
	}
	res, err := d.db.Exec(`
		INSERT INTO clinical_pathway (tag, name, medicine_branch, status, details_json)
		VALUES (?, ?, ?, ?, ?)`,
		pathway.Tag, pathway.Name, pathway.MedicineBranch, pathway.Status.String(), detailsJS)
	if err != nil {
		return err
	}
	pathway.ID, err = res.LastInsertId()
	return err
}

func (d *dataService) PathwayMenu() (*common.PathwayMenu, error) {
	var js []byte
	row := d.db.QueryRow(`
		SELECT json
		FROM clinical_pathway_menu
		WHERE status = ?
		ORDER BY created DESC
		LIMIT 1`, StatusActive)
	if err := row.Scan(&js); err == sql.ErrNoRows {
		return nil, ErrNotFound("clinical_pathway_menu")
	} else if err != nil {
		return nil, err
	}
	menu := &common.PathwayMenu{}
	return menu, json.Unmarshal(js, menu)
}

func (d *dataService) UpdatePathway(id int64, update *PathwayUpdate) error {
	args := dbutil.MySQLVarArgs()

	if update.Name != nil {
		if *update.Name == "" {
			return fmt.Errorf("api.UpdatePathway: pathway name may not be blank")
		}
		args.Append("name", *update.Name)
	}
	if update.Details != nil {
		js, err := json.Marshal(update.Details)
		if err != nil {
			return err
		}
		args.Append("details_json", js)
	}

	if args.IsEmpty() {
		return nil
	}

	_, err := d.db.Exec(`UPDATE clinical_pathway SET `+args.Columns()+` WHERE id = ?`, append(args.Values(), id)...)
	return err
}

func (d *dataService) UpdatePathwayMenu(menu *common.PathwayMenu) error {
	js, err := json.Marshal(menu)
	if err != nil {
		return err
	}
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		`UPDATE clinical_pathway_menu SET status = ? WHERE status = ?`,
		StatusInactive, StatusActive)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO clinical_pathway_menu (json, status, created)
		VALUES (?, ?, ?)`, js, StatusActive, time.Now().UTC())
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (d *dataService) CreatePathwaySTP(pathwayTag string, stp []byte) error {
	_, err := d.db.Exec(`
		UPDATE clinical_pathway
		SET stp_json = ?
		WHERE tag = ?`, stp, pathwayTag)
	return err
}

func (d *dataService) PathwaySTP(pathwayTag string) ([]byte, error) {
	var stp []byte
	if err := d.db.QueryRow(`
		SELECT stp_json
		FROM clinical_pathway
		WHERE tag = ?`, pathwayTag).Scan(&stp); err == sql.ErrNoRows {
		return nil, ErrNotFound("pathway_stp")
	} else if err != nil {
		return nil, err
	}

	return stp, nil
}

func scanPathway(opts PathwayOption, row scannable) (*common.Pathway, error) {
	p := &common.Pathway{}
	if opts&POWithDetails == 0 {
		err := row.Scan(&p.ID, &p.Tag, &p.Name, &p.MedicineBranch, &p.Status)
		if err == sql.ErrNoRows {
			return nil, ErrNotFound("clinical_pathway")
		} else if err != nil {
			return nil, err
		}
	} else {
		err := row.Scan(&p.ID, &p.Tag, &p.Name, &p.MedicineBranch, &p.Status, pathwayDetails{&p.Details})
		if err == sql.ErrNoRows {
			return nil, ErrNotFound("clinical_pathway")
		} else if err != nil {
			return nil, err
		}
	}
	return p, nil
}

type pathwayDetails struct {
	details **common.PathwayDetails
}

func (pd pathwayDetails) Scan(src interface{}) error {
	if src == nil {
		*pd.details = nil
		return nil
	}
	if s, ok := src.([]byte); ok {
		return json.Unmarshal(s, pd.details)
	}
	return fmt.Errorf("unable to scan type %T into PathwayDetails", src)
}
