package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sprucehealth/backend/Godeps/_workspace/src/github.com/SpruceHealth/schema"
	"github.com/sprucehealth/backend/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/tagging"
	"github.com/sprucehealth/backend/tagging/model"
	"github.com/sprucehealth/backend/tagging/response"
	"github.com/sprucehealth/backend/www"
)

type tagHandler struct {
	taggingClient tagging.Client
}

type tagGETRequest struct {
	Text   string `schema:"text"`
	Common bool   `json:"common"`
}

type tagGETResponse struct {
	Tags []*response.Tag `json:"tags"`
}

type tagPOSTRequest struct {
	Text   string `json:"text"`
	Common bool   `schema:"common"`
}

type tagPOSTResponse struct {
	ID int64 `json:"id"`
}

type tagPUTRequest struct {
	ID     int64 `json:"id,string"`
	Common bool  `json:"common"`
}

type tagDELETERequest struct {
	ID int64 `schema:"id,required"`
}

func newTagHandler(taggingClient tagging.Client) httputil.ContextHandler {
	return httputil.ContextSupportedMethods(&tagHandler{taggingClient: taggingClient}, httputil.Get, httputil.Put, httputil.Post, httputil.Delete)
}

func (h *tagHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "DELETE":
		req, err := h.parseDELETERequest(r)
		if err != nil {
			www.APIBadRequestError(w, r, err.Error())
			return
		}
		h.serveDELETE(w, r, req)
	case "POST":
		req, err := h.parsePOSTRequest(r)
		if err != nil {
			www.APIBadRequestError(w, r, err.Error())
			return
		}
		h.servePOST(w, r, req)
	case "PUT":
		req, err := h.parsePUTRequest(r)
		if err != nil {
			www.APIBadRequestError(w, r, err.Error())
			return
		}
		h.servePUT(w, r, req)
	case "GET":
		req, err := h.parseGETRequest(r)
		if err != nil {
			www.APIBadRequestError(w, r, err.Error())
			return
		}
		h.serveGET(w, r, req)
	}
}

func (h *tagHandler) parseGETRequest(r *http.Request) (*tagGETRequest, error) {
	rd := &tagGETRequest{}
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("Unable to parse input parameters: %s", err)
	}
	if err := schema.NewDecoder().Decode(rd, r.Form); err != nil {
		return nil, fmt.Errorf("Unable to parse input parameters: %s", err)
	}

	if rd.Text != "" && rd.Common == true {
		return nil, errors.New("The admin API only supports GET queries for exact tag text or all common tags. The text and common paramters may not be combined.")
	}
	return rd, nil
}

func (h *tagHandler) serveGET(w http.ResponseWriter, r *http.Request, req *tagGETRequest) {
	tags := make([]*response.Tag, 0, 1)
	var err error
	if !req.Common {
		if tag, err := h.taggingClient.TagFromText(req.Text); err != nil && !api.IsErrNotFound(err) {
			www.APIInternalError(w, r, err)
			return
		} else if !api.IsErrNotFound(err) {
			tags = append(tags, tag)
		}
	} else {
		ops := tagging.TONone
		if req.Common {
			ops = tagging.TOCommonOnly
		}
		tags, err = h.taggingClient.TagsFromText([]string{}, ops)
		if err != nil {
			www.APIInternalError(w, r, err)
			return
		}
	}

	httputil.JSONResponse(w, http.StatusOK, &tagGETResponse{
		Tags: tags,
	})
}

func (h *tagHandler) parsePOSTRequest(r *http.Request) (*tagPOSTRequest, error) {
	rd := &tagPOSTRequest{}
	if err := json.NewDecoder(r.Body).Decode(&rd); err != nil {
		return nil, fmt.Errorf("Unable to parse input parameters: %s", err)
	}

	if rd.Text == "" {
		return nil, errors.New("text required")
	}
	return rd, nil
}

func (h *tagHandler) servePOST(w http.ResponseWriter, r *http.Request, req *tagPOSTRequest) {
	id, err := h.taggingClient.InsertTag(&model.Tag{
		Text:   req.Text,
		Common: req.Common,
	})
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}
	httputil.JSONResponse(w, http.StatusOK, &tagPOSTResponse{
		ID: id,
	})
}

func (h *tagHandler) parsePUTRequest(r *http.Request) (*tagPUTRequest, error) {
	rd := &tagPUTRequest{}
	if err := json.NewDecoder(r.Body).Decode(&rd); err != nil {
		return nil, fmt.Errorf("Unable to parse input parameters: %s", err)
	}

	if rd.ID == 0 {
		return nil, errors.New("id required")
	}
	return rd, nil
}

func (h *tagHandler) servePUT(w http.ResponseWriter, r *http.Request, req *tagPUTRequest) {
	if err := h.taggingClient.UpdateTag(&model.TagUpdate{
		ID:     req.ID,
		Common: req.Common,
	}); err != nil {
		www.APIInternalError(w, r, err)
		return
	}
	httputil.JSONResponse(w, http.StatusOK, struct{}{})
}

func (h *tagHandler) parseDELETERequest(r *http.Request) (*tagDELETERequest, error) {
	rd := &tagDELETERequest{}
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("Unable to parse input parameters: %s", err)
	}
	if err := schema.NewDecoder().Decode(rd, r.Form); err != nil {
		return nil, fmt.Errorf("Unable to parse input parameters: %s", err)
	}
	return rd, nil
}

func (h *tagHandler) serveDELETE(w http.ResponseWriter, r *http.Request, req *tagDELETERequest) {
	if _, err := h.taggingClient.DeleteTag(req.ID); err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	httputil.JSONResponse(w, http.StatusOK, struct{}{})
}
