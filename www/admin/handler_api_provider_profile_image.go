package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/app_url"
	"github.com/sprucehealth/backend/audit"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/libs/mux"
	"github.com/sprucehealth/backend/libs/storage"
	"github.com/sprucehealth/backend/www"
	"golang.org/x/net/context"
)

type providerProfileImageAPIHandler struct {
	dataAPI    api.DataAPI
	imageStore storage.Store
	apiDomain  string
}

func newProviderProfileImageAPIHandler(dataAPI api.DataAPI, imageStore storage.Store, apiDomain string) httputil.ContextHandler {
	return httputil.SupportedMethods(&providerProfileImageAPIHandler{
		dataAPI:    dataAPI,
		imageStore: imageStore,
		apiDomain:  apiDomain,
	}, httputil.Get, httputil.Put)
}

func (h *providerProfileImageAPIHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	doctorID, err := strconv.ParseInt(mux.Vars(ctx)["id"], 10, 64)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	var imageSuffix string
	profileImageType := mux.Vars(ctx)["type"]
	switch profileImageType {
	case "thumbnail":
		// Note: for legacy reasons (when we used to have small and large thumbnails), continuing to upload
		// the thumbnail image with the large suffix
		imageSuffix = "large"
	case "hero":
		imageSuffix = "hero"
	default:
		www.APINotFound(w, r)
		return
	}

	doctor, err := h.dataAPI.GetDoctorFromID(doctorID)
	if api.IsErrNotFound(err) {
		www.APINotFound(w, r)
		return
	} else if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	account := www.MustCtxAccount(ctx)

	if r.Method == httputil.Put {
		audit.LogAction(account.ID, "AdminAPI", "UpdateProviderThumbnail", map[string]interface{}{"doctor_id": doctorID, "type": profileImageType})

		if err := r.ParseMultipartForm(maxMemory); err != nil {
			www.APIInternalError(w, r, err)
			return
		}

		file, head, err := r.FormFile("profile_image")
		if err != nil {
			www.APIInternalError(w, r, err)
			return
		}
		defer file.Close()

		size, err := common.SeekerSize(file)
		if err != nil {
			www.APIInternalError(w, r, err)
			return
		}

		meta := map[string]string{
			"X-Amz-Meta-Original-Name": head.Filename,
		}
		storeID, err := h.imageStore.PutReader(fmt.Sprintf("doctor_%d_%s", doctorID, imageSuffix), file, size, "", meta)
		if err != nil {
			www.APIInternalError(w, r, err)
			return
		}

		update := &api.DoctorUpdate{}
		switch profileImageType {
		case "thumbnail":
			update.LargeThumbnailID = &storeID
		case "hero":
			update.HeroImageID = &storeID
		}
		if err := h.dataAPI.UpdateDoctor(doctorID, update); err != nil {
			www.APIInternalError(w, r, err)
		}

		httputil.JSONResponse(w, http.StatusOK, nil)
		return
	}

	audit.LogAction(account.ID, "AdminAPI", "GetProviderThumbnail", map[string]interface{}{"doctor_id": doctorID, "type": profileImageType})

	var url string
	role := api.RoleDoctor
	if doctor.IsCC {
		role = api.RoleCC
	}
	switch profileImageType {
	case "thumbnail":
		url = app_url.ThumbnailURL(h.apiDomain, role, doctor.ID.Int64())
	case "hero":
		url = app_url.HeroImageURL(h.apiDomain, role, doctor.ID.Int64())
	}
	if url == "" {
		www.APINotFound(w, r)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}
