package admin

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"context"

	"github.com/sprucehealth/backend/cmd/svc/restapi/audit"
	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/libs/httputil"
)

const maxAnalyticsRows = 10000

type analyticsRequest struct {
	Query  string        `json:"query"`
	Params []interface{} `json:"params"`
}

type analyticsResponse struct {
	Error string          `json:"error,omitempty"`
	Cols  []string        `json:"cols,omitempty"`
	Rows  [][]interface{} `json:"rows,omitempty"`
}

type analyticsQueryAPIHandler struct {
	db *sql.DB
}

func newAnalyticsQueryAPIHandler(db *sql.DB) httputil.ContextHandler {
	return httputil.SupportedMethods(&analyticsQueryAPIHandler{
		db: db,
	}, httputil.Post)
}

func (h *analyticsQueryAPIHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req analyticsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	account := www.MustCtxAccount(ctx)
	audit.LogAction(account.ID, "AdminAPI", "AnalyticsQuery", map[string]interface{}{
		"query":  req.Query,
		"params": req.Params,
	})

	runAnalyticsQuery(w, r, h.db, req.Query, req.Params)
}

func runAnalyticsQuery(w http.ResponseWriter, r *http.Request, db *sql.DB, query string, params []interface{}) {
	rows, err := db.Query(query, params...)
	if err != nil {
		// TODO: This is super janky, but there's something either wrong with Redshift, the Postgres driver,
		// or the sql package that causes the next query to fail (causing a panic) following a bad query.
		// To contain this execute a query and recover which seems to fix it. Need to figure out what's going on,
		// but for now this "works"
		func() {
			defer func() {
				_ = recover()
			}()
			var x int
			db.QueryRow("SELECT 1").Scan(&x)
		}()
		httputil.JSONResponse(w, http.StatusOK, &analyticsResponse{Error: err.Error()})
		return
	}
	defer rows.Close()

	res := &analyticsResponse{}
	res.Cols, err = rows.Columns()
	if err != nil {
		httputil.JSONResponse(w, http.StatusOK, &analyticsResponse{Error: err.Error()})
		return
	}
	valPtrs := make([]interface{}, len(res.Cols))
	for rows.Next() {
		// rows.Scan requires ptrs to values so give it pointers to interfaces. This
		// feels terrible and one of the only places one will see pointers to interfaces,
		// but I can't think of a better way to do it.
		vals := make([]interface{}, len(res.Cols))
		for i := 0; i < len(vals); i++ {
			valPtrs[i] = &vals[i]
		}
		if err := rows.Scan(valPtrs...); err != nil {
			httputil.JSONResponse(w, http.StatusOK, &analyticsResponse{Error: err.Error()})
			return
		}

		for i, v := range vals {
			switch x := v.(type) {
			case []byte:
				vals[i] = string(x)
			}
		}

		res.Rows = append(res.Rows, vals)

		if len(res.Rows) > maxAnalyticsRows {
			break
		}
	}

	if err := rows.Err(); err != nil {
		httputil.JSONResponse(w, http.StatusOK, &analyticsResponse{Error: err.Error()})
		return
	}

	httputil.JSONResponse(w, http.StatusOK, res)
}
