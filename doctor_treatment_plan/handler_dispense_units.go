package doctor_treatment_plan

import (
	"net/http"

	"github.com/sprucehealth/backend/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/libs/httputil"
)

type medicationDispenseUnitsHandler struct {
	dataAPI api.DataAPI
}

func NewMedicationDispenseUnitsHandler(dataAPI api.DataAPI) httputil.ContextHandler {
	return httputil.SupportedMethods(
		apiservice.AuthorizationRequired(
			&medicationDispenseUnitsHandler{
				dataAPI: dataAPI,
			}), httputil.Get)
}

type MedicationDispenseUnitsResponse struct {
	DispenseUnits []*MedicationDispenseUnitItem `json:"dispense_units"`
}

type MedicationDispenseUnitItem struct {
	ID   int64  `json:"id,string"`
	Text string `json:"text"`
}

func (m *medicationDispenseUnitsHandler) IsAuthorized(ctx context.Context, r *http.Request) (bool, error) {
	if apiservice.MustCtxAccount(ctx).Role != api.RoleDoctor {
		return false, apiservice.NewAccessForbiddenError()
	}

	return true, nil
}

func (m *medicationDispenseUnitsHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	dispenseUnitIDs, dispenseUnits, err := m.dataAPI.GetMedicationDispenseUnits(api.LanguageIDEnglish)
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}
	medicationDispenseUnitResponse := &MedicationDispenseUnitsResponse{}
	medicationDispenseUnitResponse.DispenseUnits = make([]*MedicationDispenseUnitItem, len(dispenseUnits))
	for i, dispenseUnit := range dispenseUnits {
		dispenseUnitItem := &MedicationDispenseUnitItem{
			ID:   dispenseUnitIDs[i],
			Text: dispenseUnit,
		}
		medicationDispenseUnitResponse.DispenseUnits[i] = dispenseUnitItem
	}

	httputil.JSONResponse(w, http.StatusOK, medicationDispenseUnitResponse)

}
