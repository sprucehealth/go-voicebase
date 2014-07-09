package admin

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/libs/storage"
	"github.com/sprucehealth/backend/third_party/github.com/gorilla/mux"
	"github.com/sprucehealth/backend/www"
)

type doctorAttrDownloadHandler struct {
	router  *mux.Router
	dataAPI api.DataAPI
	store   storage.Store
}

func NewDoctorAttrDownloadHandler(router *mux.Router, dataAPI api.DataAPI, store storage.Store) http.Handler {
	return www.SupportedMethodsHandler(&doctorAttrDownloadHandler{
		router:  router,
		dataAPI: dataAPI,
		store:   store,
	}, []string{"GET"})
}

func (h *doctorAttrDownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	doctorID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		println(1)
		http.NotFound(w, r)
		return
	}

	attrName := vars["attr"]
	attr, err := h.dataAPI.DoctorAttributes(doctorID, []string{attrName})
	if err != nil {
		www.InternalServerError(w, r, err)
		return
	}
	if len(attr) == 0 {
		println(2)
		http.NotFound(w, r)
		return
	}

	rc, headers, err := h.store.GetReader(attr[attrName])
	if err != nil {
		println(3)
		http.NotFound(w, r)
		return
	}
	defer rc.Close()

	hd := w.Header()
	hd.Set("Content-Type", headers.Get("Content-Type"))
	hd.Set("Content-Length", headers.Get("Content-Length"))
	if fn := headers.Get("X-Amz-Meta-Original-Name"); fn != "" {
		hd.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fn))
	}
	io.Copy(w, rc)
}
