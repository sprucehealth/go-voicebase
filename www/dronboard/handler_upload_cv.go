package dronboard

import (
	"fmt"
	"html/template"
	"net/http"
	"path"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/storage"
	"github.com/sprucehealth/backend/third_party/github.com/gorilla/context"
	"github.com/sprucehealth/backend/third_party/github.com/gorilla/mux"
	"github.com/sprucehealth/backend/www"
)

const maxMemory = 5 * 1024 * 1024

type uploadHandler struct {
	router   *mux.Router
	dataAPI  api.DataAPI
	authAPI  api.AuthAPI
	store    storage.Store
	attrName string
	fileTag  string
	title    string
	nextURL  string
}

func NewUploadCVHandler(router *mux.Router, dataAPI api.DataAPI, store storage.Store) http.Handler {
	return www.SupportedMethodsHandler(&uploadHandler{
		router:   router,
		dataAPI:  dataAPI,
		store:    store,
		attrName: api.AttrCVFile,
		fileTag:  "cv",
		title:    "Upload CV / Résumé",
		nextURL:  "doctor-register-upload-license",
	}, []string{"GET", "POST"})
}

func (h *uploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		account := context.Get(r, www.CKAccount).(*common.Account)
		doctorID, err := h.dataAPI.GetDoctorIdFromAccountId(account.ID)
		if err != nil {
			www.InternalServerError(w, r, err)
			return
		}

		if err := r.ParseMultipartForm(maxMemory); err != nil {
			www.InternalServerError(w, r, err)
			return
		}

		file, fileHandler, err := r.FormFile("File")
		if err != nil {
			www.InternalServerError(w, r, err)
			return
		}
		defer file.Close()

		headers := http.Header{
			"Content-Type":  []string{fileHandler.Header.Get("Content-Type")},
			"Original-Name": []string{fileHandler.Filename},
		}

		size, err := common.SeekerSize(file)
		if err != nil {
			www.InternalServerError(w, r, err)
			return
		}

		fileID, err := h.store.PutReader(fmt.Sprintf("doctor-%d-%s%s", doctorID, h.fileTag, path.Ext(fileHandler.Filename)), file, size, headers)
		if err != nil {
			www.InternalServerError(w, r, err)
		}

		// Delete the old file if it's already been uploaded
		attr, err := h.dataAPI.DoctorAttributes(doctorID, []string{h.attrName})
		if err != nil {
			golog.Errorf("Failed to get doctor attributes for %s file: %s", h.fileTag, err.Error())
		} else if u := attr[h.attrName]; u != "" {
			if err := h.store.Delete(u); err != nil {
				golog.Errorf("Failed to delete old %s: %s", h.fileTag, err.Error())
			}
		}

		if err := h.dataAPI.UpdateDoctorAttributes(doctorID, map[string]string{h.attrName: fileID}); err != nil {
			www.InternalServerError(w, r, err)
			return
		}

		if u, err := h.router.Get(h.nextURL).URLPath(); err != nil {
			www.InternalServerError(w, r, err)
		} else {
			http.Redirect(w, r, u.String(), http.StatusSeeOther)
		}
		return
	}

	www.TemplateResponse(w, http.StatusOK, uploadTemplate, &www.BaseTemplateContext{
		Title: template.HTML(template.HTMLEscapeString(h.title) + " | Doctor Registration | Spruce"),
		SubContext: &uploadTemplateContext{
			Title: h.title,
		},
	})
}
