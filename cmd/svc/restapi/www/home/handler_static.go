package home

import (
	"html"
	"html/template"
	"net/http"
	"strings"

	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/environment"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/libs/mux"
)

type staticHandler struct {
	router   *mux.Router
	title    string
	template *template.Template
	ctx      interface{}
}

type homeContext struct {
	NoBaseHeader bool
	ExperimentID string
	SubContext   interface{}
}

func newStaticHandler(router *mux.Router, templateLoader *www.TemplateLoader, tmpl, title string, ctxFun func() interface{}) http.Handler {
	var ctx interface{}
	if ctxFun != nil {
		ctx = ctxFun()
	}
	return httputil.SupportedMethods(&staticHandler{
		router: router,
		title:  title,
		template: templateLoader.MustLoadTemplate(tmpl, "home/base.html", map[string]interface{}{
			"htmlize": func(text string) template.HTML {
				text = strings.TrimSpace(text)
				paragraphs := strings.Split(text, "\n\n")
				for i, p := range paragraphs {
					paragraphs[i] = "<p>" + strings.Replace(template.HTMLEscapeString(p), "\n", "<br>\n", -1) + "</p>"
				}
				return template.HTML(strings.Join(paragraphs, "\n"))
			},
		}),
		ctx: ctx,
	}, httputil.Get)
}

func (h *staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	www.TemplateResponse(w, http.StatusOK, h.template, &www.BaseTemplateContext{
		Environment: environment.GetCurrent(),
		Title:       template.HTML(html.EscapeString(h.title)),
		SubContext: &homeContext{
			SubContext: h.ctx,
		},
	})
}
