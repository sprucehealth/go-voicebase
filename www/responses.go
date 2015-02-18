package www

import (
	"html/template"
	"io"
	"net/http"

	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/httputil"
)

type Template interface {
	Execute(io.Writer, interface{}) error
}

const (
	HTMLContentType = "text/html; charset=utf-8"
)

// TODO: make this internal and more informative
var errorTemplate = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>{{.Title}}</title>
</head>
<body>
	<h1>{{.Title}}</h1>
	{{.Message}}
</body>
</html>
`))

type errorContext struct {
	Title   string
	Message string
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type APIErrorResponse struct {
	Error APIError `json:"error"`
}

func BadRequestError(w http.ResponseWriter, r *http.Request, err error) {
	golog.LogDepthf(1, golog.WARN, err.Error())
	TemplateResponse(w, http.StatusBadRequest, errorTemplate, &errorContext{Title: "Bad Request"})
}

func InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	golog.LogDepthf(1, golog.ERR, err.Error())
	TemplateResponse(w, http.StatusInternalServerError, errorTemplate, &errorContext{Title: "Internal Server Error"})
}

func TemplateResponse(w http.ResponseWriter, code int, tmpl Template, ctx interface{}) {
	w.Header().Set("Content-Type", HTMLContentType)
	w.WriteHeader(code)
	if err := tmpl.Execute(w, ctx); err != nil {
		golog.LogDepthf(1, golog.ERR, "Failed to render template %+v: %s", tmpl, err.Error())
	}
}

func APIInternalError(w http.ResponseWriter, r *http.Request, err error) {
	golog.LogDepthf(1, golog.ERR, err.Error())
	httputil.JSONResponse(w, http.StatusInternalServerError, &APIErrorResponse{APIError{Message: "Internal server error"}})
}

func APIBadRequestError(w http.ResponseWriter, r *http.Request, msg string) {
	httputil.JSONResponse(w, http.StatusBadRequest, &APIErrorResponse{APIError{Message: msg}})
}

func APINotFound(w http.ResponseWriter, r *http.Request) {
	httputil.JSONResponse(w, http.StatusNotFound, &APIErrorResponse{APIError{Message: "Not found"}})
}

func APIForbidden(w http.ResponseWriter, r *http.Request) {
	httputil.JSONResponse(w, http.StatusForbidden, &APIErrorResponse{APIError{Message: "Access not allowed"}})
}
