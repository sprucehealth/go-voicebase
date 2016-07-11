package cost

import (
	"net/http"

	"github.com/sprucehealth/backend/cmd/svc/restapi/analytics"
	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice"
	"github.com/sprucehealth/backend/libs/cfg"
	"github.com/sprucehealth/backend/libs/httputil"
	"golang.org/x/net/context"
)

type costHandler struct {
	dataAPI         api.DataAPI
	analyticsLogger analytics.Logger
	cfgStore        cfg.Store
}

type displayLineItem struct {
	Description string `json:"description"`
	Value       string `json:"value"`
	ChargeValue string `json:"charge_value"`
	Currency    string `json:"currency"`
}

type costResponse struct {
	LineItems []*displayLineItem `json:"line_items"`
	Total     *displayLineItem   `json:"total"`
	IsFree    bool               `json:"is_free"`
}

// NewCostHandler returns an initialized instance of costHandler
func NewCostHandler(dataAPI api.DataAPI, analyticsLogger analytics.Logger, cfgStore cfg.Store) httputil.ContextHandler {
	return httputil.SupportedMethods(
		apiservice.SupportedRoles(
			apiservice.NoAuthorizationRequired(&costHandler{
				dataAPI:         dataAPI,
				analyticsLogger: analyticsLogger,
				cfgStore:        cfgStore,
			}), api.RolePatient), httputil.Get)
}

func (c *costHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	accountID := apiservice.MustCtxAccount(ctx).ID

	skuType := r.FormValue("item_type")
	if skuType == "" {
		apiservice.WriteValidationError(ctx, "item_type required", w, r)
		return
	}

	costBreakdown, err := totalCostForItems([]string{skuType}, accountID, false, c.dataAPI, c.analyticsLogger, c.cfgStore)
	if api.IsErrNotFound(err) {
		apiservice.WriteResourceNotFoundError(ctx, "cost not found", w, r)
		return
	} else if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	response := costResponse{
		Total: &displayLineItem{
			Value:       costBreakdown.TotalCost.String(),
			Description: "Total",
			ChargeValue: costBreakdown.TotalCost.Charge(),
			Currency:    costBreakdown.TotalCost.Currency,
		},
	}

	for _, lItem := range costBreakdown.LineItems {
		response.LineItems = append(response.LineItems, &displayLineItem{
			Description: lItem.Description,
			Value:       lItem.Cost.String(),
			ChargeValue: lItem.Cost.Charge(),
			Currency:    lItem.Cost.Currency,
		})
	}

	// indicate to the client whether or not cost is free so that
	// client can leverage this information without having to parse the cost
	response.IsFree = costBreakdown.TotalCost.Amount == 0

	httputil.JSONResponse(w, http.StatusOK, response)
}