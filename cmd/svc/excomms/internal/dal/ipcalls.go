package dal

import (
	"database/sql"
	"github.com/sprucehealth/backend/cmd/svc/excomms/internal/models"
	"github.com/sprucehealth/backend/libs/dbutil"
	"github.com/sprucehealth/backend/libs/errors"
	"golang.org/x/net/context"
)

// QueryOption is an optional that can be provided to a DAL function
type QueryOption int

const (
	// ForUpdate locks the queried rows for update
	ForUpdate QueryOption = iota + 1
)

type queryOptions []QueryOption

func (qos queryOptions) Has(opt QueryOption) bool {
	for _, o := range qos {
		if o == opt {
			return true
		}
	}
	return false
}

func (d *dal) CreateIPCall(ctx context.Context, call *models.IPCall) error {
	if len(call.Participants) < 2 {
		return errors.Trace(errors.New("IPCall requires at least 2 participants"))
	}
	if call.Type == "" {
		return errors.Trace(errors.New("IPCall type required"))
	}

	var err error
	call.ID, err = models.NewIPCallID()
	if err != nil {
		return errors.Trace(err)
	}
	call.Initiated = d.clk.Now()
	call.Pending = true

	tx, err := d.db.Begin()
	if err != nil {
		return errors.Trace(err)
	}

	_, err = tx.Exec(`INSERT INTO ipcall (id, type, pending, initiated) VALUES (?, ?, ?, ?)`,
		call.ID, call.Type, call.Pending, call.Initiated)
	if err != nil {
		tx.Rollback()
		return errors.Trace(err)
	}

	for _, p := range call.Participants {
		if p.AccountID == "" {
			return errors.Trace(errors.New("IPCallParticipant account ID required"))
		}
		if p.EntityID == "" {
			return errors.Trace(errors.New("IPCallParticipant entity ID required"))
		}
		if p.Identity == "" {
			return errors.Trace(errors.New("IPCallParticipant identity required"))
		}
		if p.Role == "" {
			return errors.Trace(errors.New("IPCallParticipant role required"))
		}
		if p.State == "" {
			return errors.Trace(errors.New("IPCallParticipant state required"))
		}
		_, err := tx.Exec(`
			INSERT INTO ipcall_participant
				(ipcall_id, account_id, entity_id, identity, role, state)
			VALUES (?, ?, ?, ?, ?, ?)`, call.ID, p.AccountID, p.EntityID, p.Identity, p.Role, p.State)
		if err != nil {
			tx.Rollback()
			return errors.Trace(err)
		}
	}

	return errors.Trace(tx.Commit())
}

func (d *dal) IPCall(ctx context.Context, id models.IPCallID, opts ...QueryOption) (*models.IPCall, error) {
	forUpdate := ""
	if queryOptions(opts).Has(ForUpdate) {
		forUpdate = " FOR UPDATE"
	}

	call := &models.IPCall{ID: models.EmptyIPCallID()}
	row := d.db.QueryRow(`SELECT id, type, pending, initiated FROM ipcall WHERE id = ?`+forUpdate, id)
	if err := row.Scan(&call.ID, &call.Type, &call.Pending, &call.Initiated); err == sql.ErrNoRows {
		return nil, errors.Trace(ErrIPCallNotFound)
	} else if err != nil {
		return nil, errors.Trace(err)
	}

	rows, err := d.db.Query(`
		SELECT account_id, entity_id, identity, role, state
		FROM ipcall_participant
		WHERE ipcall_id = ?`+forUpdate, call.ID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer rows.Close()
	for rows.Next() {
		cp := &models.IPCallParticipant{}
		if err := rows.Scan(&cp.AccountID, &cp.EntityID, &cp.Identity, &cp.Role, &cp.State); err != nil {
			return nil, errors.Trace(err)
		}
		call.Participants = append(call.Participants, cp)
	}
	return call, errors.Trace(rows.Err())
}

func (d *dal) PendingIPCallsForAccount(ctx context.Context, accountID string) ([]*models.IPCall, error) {
	rows, err := d.db.Query(`
		SELECT c.id, c.type, c.pending, c.initiated
		FROM ipcall_participant cp
		INNER JOIN ipcall c ON c.id = cp.ipcall_id
		WHERE cp.account_id = ? AND pending = ?`, accountID, true)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer rows.Close()
	var calls []*models.IPCall
	for rows.Next() {
		c := &models.IPCall{ID: models.EmptyIPCallID()}
		if err := rows.Scan(&c.ID, &c.Type, &c.Pending, &c.Initiated); err != nil {
			return nil, errors.Trace(err)
		}
		calls = append(calls, c)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Trace(err)
	}

	if len(calls) == 0 {
		return calls, nil
	}

	// Query participants

	callIDs := make([]interface{}, len(calls))
	for i, c := range calls {
		callIDs[i] = c.ID
	}
	rows, err = d.db.Query(`
		SELECT ipcall_id, account_id, entity_id, identity, role, state
		FROM ipcall_participant
		WHERE ipcall_id IN (`+dbutil.MySQLArgs(len(callIDs))+`)`,
		callIDs...)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer rows.Close()
	cID := models.EmptyIPCallID()
	for rows.Next() {
		cp := &models.IPCallParticipant{}
		if err := rows.Scan(&cID, &cp.AccountID, &cp.EntityID, &cp.Identity, &cp.Role, &cp.State); err != nil {
			return nil, errors.Trace(err)
		}
		// The list of calls should generally only have 1 item so this should be plenty efficient
		for _, c := range calls {
			if c.ID.Val == cID.Val {
				c.Participants = append(c.Participants, cp)
				continue
			}
		}
	}
	return calls, errors.Trace(rows.Err())
}

func (d *dal) UpdateIPCall(ctx context.Context, callID models.IPCallID, pending bool) error {
	_, err := d.db.Exec(`UPDATE ipcall SET pending = ? WHERE id = ?`, pending, callID)
	return errors.Trace(err)
}

func (d *dal) UpdateIPCallParticipant(ctx context.Context, callID models.IPCallID, accountID string, state models.IPCallState) error {
	_, err := d.db.Exec(`UPDATE ipcall_participant SET state = ? WHERE ipcall_id = ? AND account_id = ?`, state, callID, accountID)
	return errors.Trace(err)
}