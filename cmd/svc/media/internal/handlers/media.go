package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/sprucehealth/backend/cmd/svc/media/internal/dal"
	"github.com/sprucehealth/backend/cmd/svc/media/internal/mime"
	"github.com/sprucehealth/backend/cmd/svc/media/internal/service"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/libs/media"
	"github.com/sprucehealth/backend/libs/mux"

	"golang.org/x/net/context"
)

type mediaHandler struct {
	svc service.Service
}

const contentTypeHeader = "Content-Type"

type mediaPOSTResponse struct {
	MediaID  string `json:"media_id"`
	MIMEType string `json:"mimetype"`
}

func (h *mediaHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case httputil.Get:
		h.serveGET(ctx, w, r)
	case httputil.Post:
		h.servePOST(ctx, w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *mediaHandler) servePOST(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	file, mimeType, err := parseMultiPartMedia("media", r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	// Attempt to read out a thumbnail for the media, but if anything happens drop it
	thumbFile, tType, err := parseMultiPartMedia("thumbnail", r)
	if err != nil {
		thumbFile = nil
		tType = nil
	} else if tType.Type != "image" {
		http.Error(w, fmt.Sprintf("Media type %s is not valid for thumbnails", tType.String()), http.StatusBadRequest)
		return
	} else {
		defer thumbFile.Close()
	}

	meta, err := h.svc.PutMedia(ctx, file, mimeType, thumbFile)
	if err == service.ErrUnsupportedContentType {
		http.Error(w, err.Error()+" - "+mimeType.String(), http.StatusBadRequest)
		return
	} else if err != nil {
		internalError(w, err)
		return
	}
	httputil.JSONResponse(w, http.StatusOK, &mediaPOSTResponse{
		MediaID:  meta.MediaID.String(),
		MIMEType: meta.MIMEType,
	})
}

func parseMultiPartMedia(formKey string, r *http.Request) (multipart.File, *mime.Type, error) {
	file, fHeaders, err := r.FormFile(formKey)
	if err != nil {
		return nil, nil, fmt.Errorf("Missing or invalid value for %s in parameters: %s", formKey, err)
	}
	mimeType, err := mime.ParseType(fHeaders.Header.Get(contentTypeHeader))
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to parse Content-Type for %s", formKey)
	}
	return file, mimeType, nil
}

func (h *mediaHandler) serveGET(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	mediaID, err := dal.ParseMediaID(mux.Vars(ctx)["id"])
	if err != nil {
		http.Error(w, "Cannot parse media id", http.StatusBadRequest)
		return
	}

	// For serving GET requests just redirect to the source with an expiring URL
	eURL, err := h.svc.ExpiringURL(ctx, mediaID, time.Minute*15)
	if err == dal.ErrNotFound || err == media.ErrNotFound {
		http.NotFound(w, r)
		return
	} else if err != nil {
		internalError(w, err)
		return
	}
	http.Redirect(w, r, eURL, http.StatusSeeOther)
}

func internalError(w http.ResponseWriter, err error) {
	golog.Errorf("Media: Internal Error: %s", err)
	http.Error(w, "Internal Error", http.StatusInternalServerError)
}
