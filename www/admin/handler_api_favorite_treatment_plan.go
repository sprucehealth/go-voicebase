package admin

import (
	"net/http"
	"strconv"

	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/mediastore"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/libs/mux"
	"github.com/sprucehealth/backend/responses"
	"github.com/sprucehealth/backend/www"
	"golang.org/x/net/context"
)

type ftpHandler struct {
	dataAPI    api.DataAPI
	mediaStore *mediastore.Store
}

type ftpGETResponse struct {
	FavoriteTreatmentPlan *responses.FavoriteTreatmentPlan `json:"favorite_treatment_plan"`
}

func newFTPHandler(dataAPI api.DataAPI, mediaStore *mediastore.Store) httputil.ContextHandler {
	return httputil.SupportedMethods(&ftpHandler{dataAPI: dataAPI, mediaStore: mediaStore}, httputil.Get)
}

func (h *ftpHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	ftpID, err := strconv.ParseInt(mux.Vars(ctx)["id"], 10, 64)
	if err != nil {
		www.APINotFound(w, r)
		return
	}

	switch r.Method {
	case "GET":
		h.serveGET(w, r, ftpID)
	}
}

func (h *ftpHandler) serveGET(w http.ResponseWriter, r *http.Request, ftpID int64) {
	ftp, err := h.dataAPI.FavoriteTreatmentPlan(ftpID)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	ftpr, err := responses.TransformFTPToResponse(h.dataAPI, h.mediaStore, scheduledMessageMediaExpirationDuration, ftp, "")
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, ftpGETResponse{
		FavoriteTreatmentPlan: ftpr,
	})
}
