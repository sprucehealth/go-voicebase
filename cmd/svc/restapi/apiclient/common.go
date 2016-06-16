package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice"
	"github.com/sprucehealth/backend/cmd/svc/restapi/apiservice/apipaths"
	"github.com/sprucehealth/backend/cmd/svc/restapi/appevent"
	"github.com/sprucehealth/backend/cmd/svc/restapi/handlers"
	"github.com/sprucehealth/backend/cmd/svc/restapi/passreset"
)

type Config struct {
	BaseURL    string
	AuthToken  string
	HostHeader string
}

// ResetPassword requests a password reset for the account matching the provided email
func (c *Config) ResetPassword(email string) error {
	req := &passreset.ForgotPasswordRequest{Email: email}
	return c.do("POST", apipaths.ResetPasswordURLPath, nil, req, nil, nil)
}

// AppEvent posts an app_event to the server
func (c *Config) AppEvent(action, resource string, resourceID int64) error {
	return c.do("POST", apipaths.AppEventURLPath, nil, &appevent.EventRequestData{
		Resource:   resource,
		ResourceID: resourceID,
		Action:     action,
	}, nil, nil)
}

// ListResourceGuides returns a list of resource guides broken into sections.
func (c *Config) ListResourceGuides() ([]*handlers.Section, error) {
	var res handlers.ListResponse
	err := c.do("GET", apipaths.ResourceGuidesListURLPath, nil, nil, &res, nil)
	return res.Sections, err
}

// ResourceGuide returns the decoded resource guide for the provided ID.
func (c *Config) ResourceGuide(id int64) (interface{}, error) {
	var v interface{}
	err := c.do("GET", apipaths.ResourceGuideURLPath,
		url.Values{"resource_id": []string{strconv.FormatInt(id, 10)}}, nil, &v, nil)
	return v, err
}

func (c *Config) do(method, path string, params url.Values, req, res interface{}, headers http.Header) error {
	var body io.Reader
	if req != nil {
		if r, ok := req.(io.Reader); ok {
			body = r
		} else if b, ok := req.([]byte); ok {
			body = bytes.NewReader(b)
		} else {
			if headers == nil {
				headers = http.Header{}
			}
			headers.Set("Content-Type", "application/json")
			b := &bytes.Buffer{}
			if err := json.NewEncoder(b).Encode(req); err != nil {
				return err
			}
			body = b
		}
	}

	u := c.BaseURL + path
	if len(params) != 0 {
		u += "?" + params.Encode()
	}
	httpReq, err := http.NewRequest(method, u, body)
	if err != nil {
		return err
	}
	for k, v := range headers {
		httpReq.Header[k] = v
	}
	if c.AuthToken != "" {
		httpReq.Header.Set("Authorization", "token "+c.AuthToken)
	}
	if c.HostHeader != "" {
		httpReq.Host = c.HostHeader
	}
	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpRes.Body.Close()

	switch httpRes.StatusCode {
	case http.StatusMethodNotAllowed:
		return fmt.Errorf("apiclient: method %s not allowed on endpoint '%s'", method, path)
	case http.StatusOK:
		if res != nil {
			return json.NewDecoder(httpRes.Body).Decode(res)
		}
		return nil
	}

	var e apiservice.SpruceError
	if err := json.NewDecoder(httpRes.Body).Decode(&e); err != nil {
		return fmt.Errorf("apiclient: failed to decode error on %d status code: %s", httpRes.StatusCode, err.Error())
	}
	e.HTTPStatusCode = httpRes.StatusCode
	return &e
}
