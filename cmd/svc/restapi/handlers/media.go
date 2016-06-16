package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image/jpeg"
	_ "image/png" // imported to register PNG decoder
	"io"
	"net/http"
	"time"

	"github.com/samuel/go-metrics/metrics"
	"github.com/sprucehealth/backend/cmd/svc/restapi/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice"
	"github.com/sprucehealth/backend/cmd/svc/restapi/mediastore"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/libs/imageutil"
	"github.com/sprucehealth/backend/libs/media"
	"github.com/sprucehealth/backend/libs/storage"
	"golang.org/x/net/context"
)

type mediaHandler struct {
	dataAPI            api.DataAPI
	store              *mediastore.Store
	cacheStore         storage.DeterministicStore
	expirationDuration time.Duration
	statCacheHit       *metrics.Counter
	statCacheMiss      *metrics.Counter
	statTotalLatency   metrics.Histogram
	statResizeLatency  metrics.Histogram
	statWriteLatency   metrics.Histogram
}

type uploadResponse struct {
	MediaID  int64  `json:"media_id,string"`
	PhotoID  int64  `json:"photo_id,string"`
	MediaURL string `json:"media_url"`
	PhotoURL string `json:"photo_url"`
}

type mediaRequest struct {
	MediaID    int64  `schema:"media_id,required"`
	Signature  string `schema:"sig,required"`
	ExpireTime int64  `schema:"expires,required"`
	Width      int    `schema:"width"`
	Height     int    `schema:"height"`
}

// NewMedia returns an initialized media handler.
func NewMedia(
	dataAPI api.DataAPI,
	store *mediastore.Store,
	cacheStore storage.DeterministicStore,
	expirationDuration time.Duration,
	metricsRegistry metrics.Registry,
) httputil.ContextHandler {
	h := &mediaHandler{
		dataAPI:            dataAPI,
		store:              store,
		cacheStore:         cacheStore,
		expirationDuration: expirationDuration,
		statCacheHit:       metrics.NewCounter(),
		statCacheMiss:      metrics.NewCounter(),
		statTotalLatency:   metrics.NewUnbiasedHistogram(),
		statResizeLatency:  metrics.NewUnbiasedHistogram(),
		statWriteLatency:   metrics.NewUnbiasedHistogram(),
	}
	metricsRegistry.Add("cache/hit", h.statCacheHit)
	metricsRegistry.Add("cache/miss", h.statCacheMiss)
	metricsRegistry.Add("latency/total", h.statTotalLatency)
	metricsRegistry.Add("latency/resize", h.statResizeLatency)
	metricsRegistry.Add("latency/write", h.statWriteLatency)
	return httputil.SupportedMethods(
		apiservice.RequestCacheHandler(
			apiservice.AuthorizationRequired(h)),
		httputil.Get, httputil.Post)
}

func (h *mediaHandler) IsAuthorized(ctx context.Context, r *http.Request) (bool, error) {
	requestCache := apiservice.MustCtxCache(ctx)
	switch r.Method {
	case httputil.Get:
		req := &mediaRequest{}
		if err := apiservice.DecodeRequestData(req, r); err != nil {
			return false, apiservice.NewValidationError(err.Error())
		}
		if req.ExpireTime < time.Now().UTC().Unix() {
			return false, apiservice.NewAccessForbiddenError()
		}
		if !h.store.ValidateSignature(req.MediaID, req.ExpireTime, req.Signature) {
			return false, apiservice.NewAccessForbiddenError()
		}
		requestCache[apiservice.CKRequestData] = req
	case httputil.Post:
		account, ok := apiservice.CtxAccount(ctx)
		if !ok {
			return false, apiservice.NewAccessForbiddenError()
		}
		var personID int64
		switch account.Role {
		case api.RoleDoctor, api.RoleCC:
			doctorID, err := h.dataAPI.GetDoctorIDFromAccountID(account.ID)
			if err != nil {
				return false, err
			}
			personID, err = h.dataAPI.GetPersonIDByRole(account.Role, doctorID)
			if err != nil {
				return false, err
			}
		case api.RolePatient:
			patientID, err := h.dataAPI.GetPatientIDFromAccountID(account.ID)
			if err != nil {
				return false, err
			}
			personID, err = h.dataAPI.GetPersonIDByRole(api.RolePatient, patientID.Int64())
			if err != nil {
				return false, err
			}
		default:
			return false, apiservice.NewAccessForbiddenError()
		}
		requestCache[apiservice.CKPersonID] = personID
	}
	return true, nil
}

func (h *mediaHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case httputil.Get:
		h.get(ctx, w, r)
	case httputil.Post:
		h.post(ctx, w, r)
	}
}

func copyWithHeaders(w http.ResponseWriter, r io.Reader, headers http.Header, lastModified time.Time) {
	w.Header().Set("Content-Type", headers.Get("Content-Type"))
	if cl := headers.Get("Content-Length"); cl != "" {
		w.Header().Set("Content-Length", cl)
	}
	httputil.FarFutureCacheHeaders(w.Header(), lastModified)
	io.Copy(w, r)
}

func (h *mediaHandler) get(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	req := apiservice.MustCtxCache(ctx)[apiservice.CKRequestData].(*mediaRequest)

	startTime := time.Now()

	med, err := h.dataAPI.GetMedia(req.MediaID)
	if api.IsErrNotFound(err) {
		apiservice.WriteResourceNotFoundError(ctx, "Media not found", w, r)
		return
	} else if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	// TODO: Check "If-Modified-Since"

	// Stream original image if no resizing is requested
	if req.Width <= 0 && req.Height <= 0 {
		rc, headers, err := h.store.GetReader(med.URL)
		if err != nil {
			apiservice.WriteError(ctx, err, w, r)
			return
		}
		defer rc.Close()
		copyWithHeaders(w, rc, headers, med.Uploaded)
		h.statTotalLatency.Update(time.Since(startTime).Nanoseconds() / 1e3)
		return
	}

	// Resizing is request so first check the cache

	cacheKey := fmt.Sprintf("%d-%dx%d", req.MediaID, req.Width, req.Height)

	if h.cacheStore != nil {
		rc, headers, err := h.cacheStore.GetReader(h.cacheStore.IDFromName(cacheKey))
		if err == nil {
			defer rc.Close()
			h.statCacheHit.Inc(1)
			copyWithHeaders(w, rc, headers, med.Uploaded)
			h.statTotalLatency.Update(time.Since(startTime).Nanoseconds() / 1e3)
			return
		}
	}
	h.statCacheMiss.Inc(1)

	// Not in the cache so generate the requested size

	rc, _, err := h.store.GetReader(med.URL)
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}
	defer rc.Close()

	resizeStartTime := time.Now()
	resizedImg, _, err := imageutil.ResizeImageFromReader(rc, req.Width, req.Height, nil)
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}
	h.statResizeLatency.Update(time.Since(resizeStartTime).Nanoseconds() / 1e3)

	w.Header().Set("Content-Type", "image/jpeg")
	httputil.FarFutureCacheHeaders(w.Header(), med.Uploaded)

	// TODO: Dual stream encoding to cache and response once the s3 storage
	// implements streaming multi-writer. S3 requires specifying the content-length
	// on uploads so it's necessary to use a multi-part upload when the size is
	// unknown. However, the s3 package we're using currently doesn't allow setting
	// headers multi-part uploads.

	writeStartTime := time.Now()
	buf := &bytes.Buffer{}
	if err := jpeg.Encode(buf, resizedImg, &jpeg.Options{
		Quality: imageutil.JPEGQuality,
	}); err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	// Save resized image to the cache
	go func() {
		if _, err := h.cacheStore.Put(cacheKey, buf.Bytes(), "image/jpeg", nil); err != nil {
			golog.Errorf("Failed to write resize image to cache: %s", err.Error())
		}
	}()

	w.Write(buf.Bytes())
	h.statWriteLatency.Update(time.Since(writeStartTime).Nanoseconds() / 1e3)

	h.statTotalLatency.Update(time.Since(startTime).Nanoseconds() / 1e3)
}

func (h *mediaHandler) post(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	requestCache := apiservice.MustCtxCache(ctx)
	personID := requestCache[apiservice.CKPersonID].(int64)
	file, handler, err := r.FormFile("media")
	if err != nil {
		file, handler, err = r.FormFile("photo")
		if err != nil {
			apiservice.WriteUserError(w, http.StatusBadRequest, "Missing or invalid media in parameters: "+err.Error())
			return
		}
	}
	defer file.Close()

	size, err := media.SeekerSize(file)
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	uid := make([]byte, 16)
	if _, err := rand.Read(uid); err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	name := "media-" + hex.EncodeToString(uid)
	contentType := handler.Header.Get("Content-Type")

	url, err := h.store.PutReader(name, file, size, contentType, nil)
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	id, err := h.dataAPI.AddMedia(personID, url, contentType)
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}

	signedURL, err := h.store.SignedURL(id, h.expirationDuration)
	if err != nil {
		apiservice.WriteError(ctx, err, w, r)
		return
	}
	res := &uploadResponse{
		MediaID:  id,
		PhotoID:  id,
		MediaURL: signedURL,
		PhotoURL: signedURL,
	}
	httputil.JSONResponse(w, http.StatusOK, res)
}
