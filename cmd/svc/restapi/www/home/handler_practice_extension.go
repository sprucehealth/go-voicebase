package home

import (
	"html"
	"html/template"
	"net/http"

	"context"

	"github.com/sprucehealth/backend/cmd/svc/restapi/www"
	"github.com/sprucehealth/backend/environment"
	"github.com/sprucehealth/backend/libs/httputil"
)

type practiceExtensionHandler struct {
	template *template.Template
	title    string
	ctx      interface{}
}

func newPracticeExtensionStaticHandler(templateLoader *www.TemplateLoader, template string, title string, ctxFun func() interface{}) httputil.ContextHandler {
	var ctx interface{}
	if ctxFun != nil {
		ctx = ctxFun()
	}
	return httputil.SupportedMethods(&practiceExtensionHandler{
		title:    title,
		template: templateLoader.MustLoadTemplate(template, "practice-extension/base.html", nil),
		ctx:      ctx,
	}, httputil.Get)
}

func (h *practiceExtensionHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	www.TemplateResponse(w, http.StatusOK, h.template, &struct {
		Environment string
		Title       template.HTML
		SubContext  interface{}
	}{
		Environment: environment.GetCurrent(),
		Title:       template.HTML(html.EscapeString(h.title)),
		SubContext:  h.ctx,
	})
}
