package treatment_plan

import (
	"bytes"
	"html/template"
	"io"
	"strings"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/common"
)

var templateFuncMap = map[string]interface{}{
	"renderView": func(view tpView) (template.HTML, error) {
		buf := &bytes.Buffer{}
		if err := treatmentTemplate.ExecuteTemplate(buf, view.TypeName(), view); err != nil {
			return "", err
		}
		return template.HTML(buf.String()), nil
	},
	"mapImageURL": func(url string) (string, error) {
		if strings.HasPrefix(url, "spruce:///image/") {
			// TODO: this URL should be taken from the config
			return "https://carefront-static.s3.amazonaws.com/" + url[16:], nil
		}
		return url, nil
	},
}

const templateText = `
{{define "base"}}
	<div class="treatment-plan">
		{{range .Views}}
			{{renderView .}}
		{{end}}
	</div>
{{end}}

{{define "treatment:image"}}
	{{if .ImageURL}}
		<img src="{{mapImageURL .ImageURL}}"> <!-- width="{{.ImageWidth}}" height="{{.ImageHeight}}" -->
	{{end}}
{{end}}

{{define "treatment:small_divider"}}
	<hr class="small-divider">
{{end}}

{{define "treatment:large_divider"}}
	<div class="large-divider-view">&nbsp;</div>
{{end}}

{{define "treatment:list_element"}}
	<div class="list-element content-view">
		{{if eq .ElementStyle "numbered"}}
			<div style="float:left; width:20px; text-align:right;">{{.Number}}.</div><div style="margin-left:25px;">{{.Text}}</div>
		{{else}}
			<div style="float:left; width:15px; text-align:center;">●</div><div style="margin-left:20px;">{{.Text}}</div>
		{{end}}
	</div>
{{end}}

{{define "treatment:icon_title_subtitle_view"}}
	<div class="icon-title-subtitle-view content-view">
		{{if .IconURL}}<img src="{{mapImageURL .IconURL.String}}" width="32" height="32">{{end}}
		<div class="title">{{.Title}}</div>
		<div class="subtitle">{{.Subtitle}}</div>
	</div>
{{end}}

{{define "treatment:icon_text_view"}}
	<div class="icon-text-view content-view {{.Style}}">
		{{if .IconURL}}<img src="{{mapImageURL .IconURL.String}}" width="{{.IconWidth}}" height="{{.IconHeight}}">{{end}}
		<span class="{{.TextStyle}}">{{.Text}}</span>
	</div>
{{end}}

{{define "treatment:text"}}
	<div class="text-view content-view text-view-style-{{.Style}}">
		{{.Text}}
	</div>
{{end}}

{{define "treatment:button"}}
	<div class="button-view content-view">
		<a href="{{.TapURL}}"><img src="{{mapImageURL .IconURL.String}}"> {{.Text}}</a>
	</div>
{{end}}

{{define "treatment:hero_header"}}
	<div class="hero-header">
		<h3 class="title">{{.Title}}</h3>
		<div class="subtitle">{{.Subtitle}} <span class="created-date">{{.CreatedDateText}}</span></div>
	</div>
{{end}}

{{define "treatment:card_view"}}
	<div class="treatment-card-view">
		{{range .Views}}
			{{renderView .}}
		{{end}}
	</div>
{{end}}

{{define "treatment:card_title_view"}}
	<div class="card-title-view">
		<!-- <img src="{{mapImageURL .IconURL.String}}"> -->
		<h4 class="title">{{.Title}}</h4>
	</div>
{{end}}

{{define "treatment:pharmacy"}}
	<div class="pharmacy">
		<h4>Pharmacy</h4>
		<!-- <p>{{.Text}}</p> -->
		<div class="name">{{.Pharmacy.Name}}</div>
		<div class="phone">{{.Pharmacy.Phone}}</div>
		{{with .Pharmacy.Fax}}<div class="fax">FAX: {{.}}</div>{{end}}
		<div class="address">
			<div>{{.Pharmacy.AddressLine1}}</div>
			{{with .Pharmacy.AddressLine2}}<div>{{.}}</div>{{end}}
			<div>{{.Pharmacy.City}}, {{.Pharmacy.State}}</div>
			<div>{{.Pharmacy.Postal}}</div>
		</div>
	</div>
{{end}}

{{define "treatment:prescription"}}
	<div class="prescription">
		<h4>Prescription</h4>
		<!-- <img src="{{mapImageURL .IconURL.String}}"> -->
		<div class="title">{{.Title}}</div>
		<!-- <div class="small-header-text">{{.SmallHeaderText}}</div> -->
		<div class="description">{{.Description}}</div>
	</div>
{{end}}

{{define "treatment:button_footer"}}
{{end}}
`

var treatmentTemplate *template.Template

func init() {
	treatmentTemplate = template.Must(template.New("").Funcs(templateFuncMap).Parse(templateText))
}

type rxGuideTemplateContext struct {
	Views []tpView
}

func RenderRXGuide(w io.Writer, details *common.DrugDetails, treatment *common.Treatment, treatmentPlan *common.TreatmentPlan) error {
	views, err := treatmentGuideViews(details, treatment, treatmentPlan)
	if err != nil {
		return err
	}
	return treatmentTemplate.ExecuteTemplate(w, "base", &rxGuideTemplateContext{Views: views})
}

func RenderTreatmentPlan(w io.Writer, dataAPI api.DataAPI, treatmentPlan *common.TreatmentPlan, doctor *common.Doctor, patient *common.Patient) error {
	if err := populateTreatmentPlan(dataAPI, treatmentPlan); err != nil {
		return err
	}

	res, err := treatmentPlanResponse(dataAPI, treatmentPlan, doctor, patient)
	if err != nil {
		return err
	}

	if err := treatmentTemplate.ExecuteTemplate(w, "base", &rxGuideTemplateContext{Views: res.HeaderViews}); err != nil {
		return err
	}
	if err := treatmentTemplate.ExecuteTemplate(w, "base", &rxGuideTemplateContext{Views: res.TreatmentViews}); err != nil {
		return err
	}
	if err := treatmentTemplate.ExecuteTemplate(w, "base", &rxGuideTemplateContext{Views: res.InstructionViews}); err != nil {
		return err
	}
	return nil
}
