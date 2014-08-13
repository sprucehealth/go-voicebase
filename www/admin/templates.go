package admin

import (
	"html/template"

	"github.com/sprucehealth/backend/common"
)

// var (
// 	baseTemplate              *template.Template
// 	doctorSearchTemplate      *template.Template
// 	doctorTemplate            *template.Template
// 	resourceGuideListTemplate *template.Template
// 	resourceGuideTemplate     *template.Template
// 	rxGuideListTemplate       *template.Template
// 	rxGuideTemplate           *template.Template
// )

// func init() {
// 	baseTemplate = www.MustLoadTemplate("admin/base.html", template.Must(www.BaseTemplate.Clone()))
// 	doctorSearchTemplate = www.MustLoadTemplate("admin/doctor_search.html", template.Must(baseTemplate.Clone()))
// 	doctorTemplate = www.MustLoadTemplate("admin/doctor.html", template.Must(baseTemplate.Clone()))
// 	resourceGuideListTemplate = www.MustLoadTemplate("admin/resourceguide_list.html", template.Must(baseTemplate.Clone()))
// 	resourceGuideTemplate = www.MustLoadTemplate("admin/resourceguide.html", template.Must(baseTemplate.Clone()))
// 	rxGuideListTemplate = www.MustLoadTemplate("admin/rxguide_list.html", template.Must(baseTemplate.Clone()))
// 	rxGuideTemplate = www.MustLoadTemplate("admin/rxguide.html", template.Must(baseTemplate.Clone()))
// }

type doctorSearchTemplateContext struct {
	Query   string
	Doctors []*common.DoctorSearchResult
}

type doctorTemplateContext struct {
	Doctor          *common.Doctor
	Attributes      map[string]template.HTML
	MedicalLicenses []*common.MedicalLicense
}

type resourceGuideListTemplateContext struct {
	Sections []*common.ResourceGuideSection
	Guides   map[int64][]*common.ResourceGuide
}

type resourceGuideTemplateContext struct {
	Form  *resourceGuideForm
	Error string
}

type rxGuideListTemplateContext struct {
	Drugs []*common.DrugDetails
}

type rxGuideTemplateContext struct {
	Details     *common.DrugDetails
	DetailsHTML template.HTML
}
