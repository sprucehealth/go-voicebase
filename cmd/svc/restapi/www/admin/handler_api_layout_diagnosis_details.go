package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/common"
	"github.com/sprucehealth/backend/cmd/svc/restapi/diagnosis"
	"github.com/sprucehealth/backend/cmd/svc/restapi/info_intake"
	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/libs/httputil"
	"golang.org/x/net/context"
)

type diagDetailsLayoutUploadHandler struct {
	dataAPI      api.DataAPI
	diagnosisAPI diagnosis.API
}

type diagnosisLayoutItems struct {
	Items []*diagnosisLayoutItem `json:"diagnosis_layouts"`
}

type diagnosisLayoutItem struct {
	CodeID        string            `json:"code_id"`
	LayoutVersion *encoding.Version `json:"layout_version"`
	Questions     json.RawMessage   `json:"questions"`
}

func newDiagnosisDetailsIntakeUploadHandler(dataAPI api.DataAPI, diagnosisAPI diagnosis.API) httputil.ContextHandler {
	return httputil.SupportedMethods(&diagDetailsLayoutUploadHandler{dataAPI, diagnosisAPI}, httputil.Post)
}

func (d *diagDetailsLayoutUploadHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	rd := &diagnosisLayoutItems{}
	if err := r.ParseForm(); err != nil {
		www.BadRequestError(w, r, err)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(rd); err != nil {
		www.BadRequestError(w, r, err)
		return
	}

	// ensure that the diagnosis codes exist
	codeIDs := make([]string, len(rd.Items))
	for i, item := range rd.Items {
		codeIDs[i] = item.CodeID
	}

	if res, nonExistentCodeIDs, err := d.diagnosisAPI.DoCodesExist(codeIDs); err != nil {
		www.APIInternalError(w, r, err)
		return
	} else if !res {
		www.APIBadRequestError(w, r, fmt.Sprintf("Following codes do not exist: %v", nonExistentCodeIDs))
		return
	}

	// ensure that for each of the incoming diagnosis the layout inputted is higher than the layout already
	// supported for the version
	var errors []string
	for _, item := range rd.Items {
		existingVersion, err := d.dataAPI.ActiveDiagnosisDetailsIntakeVersion(item.CodeID)
		if api.IsErrNotFound(err) {
			continue
		} else if err != nil {
			www.APIInternalError(w, r, err)
			return
		}
		if !existingVersion.LessThan(item.LayoutVersion) {
			errors = append(errors,
				fmt.Sprintf("Incoming layout version %s is less than existing layout version %s for codeID %s",
					item.LayoutVersion.String(), existingVersion.String(), item.CodeID))
		}
	}
	if len(errors) > 0 {
		www.APIBadRequestError(w, r, strings.Join(errors, "\n"))
		return
	}

	// for each layout entry, create a template, fill in the questions and then create the actual layout
	for _, item := range rd.Items {

		// unmarshal the quesitons into two separate objects so that
		// we have a copy for the template and then a copy into which to fill the
		// question information
		var qIntake []*info_intake.Question
		if err := json.Unmarshal(item.Questions, &qIntake); err != nil {
			www.APIInternalError(w, r, err)
			return
		}

		layout := diagnosis.NewQuestionIntake(qIntake)
		template := &common.DiagnosisDetailsIntake{
			CodeID:  item.CodeID,
			Version: item.LayoutVersion,
			Active:  true,
			Layout:  &layout,
		}

		if err := json.Unmarshal(item.Questions, &qIntake); err != nil {
			www.APIInternalError(w, r, err)
			return
		}

		if err := api.FillQuestions(qIntake, d.dataAPI, api.LanguageIDEnglish); err != nil {
			www.APIInternalError(w, r, err)
			return
		}

		layout = diagnosis.NewQuestionIntake(qIntake)
		info := &common.DiagnosisDetailsIntake{
			CodeID:  item.CodeID,
			Version: item.LayoutVersion,
			Active:  true,
			Layout:  &layout,
		}

		// save the template and the fleshed out object into the database
		if err := d.dataAPI.SetDiagnosisDetailsIntake(template, info); err != nil {
			www.APIInternalError(w, r, err)
			return
		}
	}
	httputil.JSONResponse(w, http.StatusOK, nil)
}
