package patient

import (
	"net/http"

	"github.com/sprucehealth/backend/address"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/surescripts"
)

type UpdateHandler struct {
	dataAPI          api.DataAPI
	addressValidator address.Validator
}

type PhoneNumber struct {
	Type   string `json:"phone_type,omitempty"`
	Number string `json:"phone"`
}

type UpdateRequest struct {
	PhoneNumbers []PhoneNumber   `json:"phone_numbers"`
	Address      *common.Address `json:"address"`
}

func (r *UpdateRequest) isZero() bool {
	return (r == nil || (len(r.PhoneNumbers) == 0 && r.Address == nil))
}

func (r *UpdateRequest) transformRequestToUpdate(dataAPI api.DataAPI, validator address.Validator) (*api.PatientUpdate, error) {
	var update api.PatientUpdate
	var err error

	if len(r.PhoneNumbers) > 0 {
		update.PhoneNumbers, err = transformPhoneNumbers(r.PhoneNumbers)
		if err != nil {
			return nil, err
		}
	}

	if r.Address != nil {
		if err := surescripts.ValidateAddress(r.Address, validator, dataAPI); err != nil {
			return nil, err
		}

		update.Address = r.Address
	}

	return &update, nil
}

func NewUpdateHandler(dataAPI api.DataAPI, addressValidator address.Validator) http.Handler {
	return httputil.SupportedMethods(
		apiservice.SupportedRoles(
			apiservice.NoAuthorizationRequired(&UpdateHandler{
				dataAPI:          dataAPI,
				addressValidator: addressValidator,
			}),
			[]string{api.RolePatient},
		), httputil.Post, httputil.Put)
}

func (h *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		apiservice.WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse input parameters: "+err.Error())
		return
	}

	req := &UpdateRequest{}
	if err := apiservice.DecodeRequestData(req, r); err != nil {
		apiservice.WriteValidationError(err.Error(), w, r)
		return
	}

	patientID, err := h.dataAPI.GetPatientIDFromAccountID(apiservice.GetContext(r).AccountID)
	if err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	// TODO: implement DELETE
	switch r.Method {
	case "POST", "PUT":
		// For now treat these the same because we don't support more than one phone number
		// for the patient which is the only this this endpoint currently supports.
		h.postOrPUT(w, r, patientID, req)
	}
}

func (h *UpdateHandler) postOrPUT(w http.ResponseWriter, r *http.Request, patientID int64, req *UpdateRequest) {

	if req.isZero() {
		apiservice.WriteJSONSuccess(w)
		return
	}

	update, err := req.transformRequestToUpdate(h.dataAPI, h.addressValidator)
	if err != nil {
		apiservice.WriteValidationError(err.Error(), w, r)
		return
	}

	if err := h.dataAPI.UpdatePatient(patientID, update, false); err != nil {
		apiservice.WriteError(err, w, r)
		return
	}

	apiservice.WriteJSONSuccess(w)
}

func transformPhoneNumbers(pn []PhoneNumber) ([]*common.PhoneNumber, error) {
	var numbers []*common.PhoneNumber
	for _, phone := range pn {
		num, err := common.ParsePhone(phone.Number)
		if err != nil {
			return nil, err
		}
		numbers = append(numbers, &common.PhoneNumber{
			Phone: num,
			Type:  phone.Type,
		})
	}
	return numbers, nil
}
