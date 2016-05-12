package admin

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/cmd/svc/restapi/erx"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/doctor_treatment_plan"
	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/libs/ptr"
	"github.com/sprucehealth/backend/pharmacy"
	"github.com/sprucehealth/backend/treatment_plan"
	"github.com/sprucehealth/backend/views"
	"github.com/sprucehealth/backend/www"
	"golang.org/x/net/context"
)

var (
	noteRXDescriptionRE          = regexp.MustCompile(`note:rx_\d+_description`)
	noteAdditionalInfoRE         = regexp.MustCompile(`note:additional_info_\d+`)
	scheduledMessageDurationRE   = regexp.MustCompile(`scheduled_message_\d+_duration`)
	scheduledMessageUnitRE       = regexp.MustCompile(`scheduled_message_\d+_unit`)
	scheduledMessageAttachmentRE = regexp.MustCompile(`scheduled_message_\d+_attachment`)
	scheduledMessageRE           = regexp.MustCompile(`scheduled_message_\d+`)
	rxNameRE                     = regexp.MustCompile(`rx_\d+_name`)
	rxDosageRE                   = regexp.MustCompile(`rx_\d+_dosage`)
	rxDispenseTypeRE             = regexp.MustCompile(`rx_\d+_dispense_type`)
	rxDispenseNumberRE           = regexp.MustCompile(`rx_\d+_dispense_number`)
	rxRefillsRE                  = regexp.MustCompile(`rx_\d+_refills`)
	rxSubstitutionsRE            = regexp.MustCompile(`rx_\d+_substitutions`)
	rxSigRE                      = regexp.MustCompile(`rx_\d+_sig`)
	rxRouteRE                    = regexp.MustCompile(`rx_\d+_route`)
	rxFormRE                     = regexp.MustCompile(`rx_\d+_form`)
	rxGenericNameRE              = regexp.MustCompile(`rx_\d+_generic_name`)
	sectionTitleRE               = regexp.MustCompile(`section_\d+_title`)
	sectionStepRE                = regexp.MustCompile(`section_\d+_step_\d+`)
	resourceGuideRE              = regexp.MustCompile(`resource_guide_\d+`)
	digitsRE                     = regexp.MustCompile(`\d+`)
)

type ftp struct {
	FrameworkTag      string
	FrameworkName     string
	SFTPName          string
	Diagnosis         string
	Note              note
	ScheduledMessages map[string]scheduledMessage
	Sections          map[string]section
	RXs               map[string]rx
	ResourceGuideTags []string
	IsSTP             bool
}

type note struct {
	Welcome              string
	ConditionDescription string
	MDRecommendation     string
	RXDescriptions       map[string]string
	AdditionalInfo       map[string]string
	Closing              string
}

func (n note) String() string {
	var rxDescription string
	for _, v := range n.RXDescriptions {
		rxDescription += v + "\n\n"
	}
	var additionalInfo string
	additionalInfoKeys := make([]string, 0, len(n.AdditionalInfo))
	for k := range n.AdditionalInfo {
		additionalInfoKeys = append(additionalInfoKeys, k)
	}
	sort.Strings(additionalInfoKeys)
	for _, k := range additionalInfoKeys {
		additionalInfo += n.AdditionalInfo[k] + "\n\n"
	}
	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s%s%s\n", n.Welcome, n.ConditionDescription, n.MDRecommendation, rxDescription, additionalInfo, n.Closing)
}

type rx struct {
	Name           string
	Dosage         string
	DispenseType   string
	DispenseNumber string
	Refills        string
	Substitutions  string
	Sig            string
}

type scheduledMessage struct {
	Duration   string
	Unit       string
	Message    string
	Attachment string
}

func (sm scheduledMessage) DurationInDays() (int, error) {
	multiplier := 1
	if sm.Unit == "weeks" {
		multiplier = 7
	} else if sm.Unit == "months" {
		multiplier = 30
	} else if sm.Unit == "years" {
		multiplier = 365
	}
	dur, err := strconv.ParseInt(sm.Duration, 10, 64)
	if err != nil {
		return 0, err
	}
	return (int(dur) * multiplier), nil
}

func (sm scheduledMessage) RequiresFollowup() bool {
	return strings.ToLower(sm.Attachment) == "yes" || strings.ToLower(sm.Attachment) == "true"
}

type section struct {
	Title string
	Steps []step
}

type step struct {
	Text  string
	Order int64
}

type treatmentPlanCSVHandler struct {
	dataAPI api.DataAPI
	erxAPI  erx.ERxAPI
}

type treatmentPlanCSVPUTRequest struct {
	ColData [][]string
}

func newFTP() *ftp {
	return &ftp{
		Note: note{
			RXDescriptions: make(map[string]string),
			AdditionalInfo: make(map[string]string),
		},
		ScheduledMessages: make(map[string]scheduledMessage),
		Sections:          make(map[string]section),
		RXs:               make(map[string]rx),
	}
}

func newTreatmentPlanCSVHandler(dataAPI api.DataAPI, erxAPI erx.ERxAPI) httputil.ContextHandler {
	return httputil.SupportedMethods(&treatmentPlanCSVHandler{dataAPI: dataAPI, erxAPI: erxAPI}, httputil.Put)
}

func (h *treatmentPlanCSVHandler) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		requestData, err := h.parsePUTRequest(ctx, r)
		if err != nil {
			www.APIBadRequestError(w, r, err.Error())
			return
		}
		h.servePUT(ctx, w, r, requestData)
	}
}

func (h *treatmentPlanCSVHandler) parsePUTRequest(ctx context.Context, r *http.Request) (*treatmentPlanCSVPUTRequest, error) {
	var err error
	rd := &treatmentPlanCSVPUTRequest{}
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		return nil, err
	}
	f, _, err := r.FormFile("csv")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rd.ColData, err = csvDataFromFile(f)
	if err != nil {
		return nil, err
	}

	return rd, nil
}

func (h *treatmentPlanCSVHandler) servePUT(ctx context.Context, w http.ResponseWriter, r *http.Request, req *treatmentPlanCSVPUTRequest) {
	threads := len(req.ColData)
	ftps, err := parseFTPs(req.ColData, threads)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	for _, v := range ftps {
		if v.FrameworkTag == "" {
			www.APIBadRequestError(w, r, fmt.Sprintf("Empty framework_tag detected. Cannot complete request"))
			return
		}
	}

	err = h.createGlobalFTPs(ftps)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}

	_, err = h.transformFTPsToSTPs(ctx, ftps, threads)
	if err != nil {
		www.APIInternalError(w, r, err)
		return
	}
}

type completedSTP struct {
	PathwayTag string
	STPJSON    []byte
}

func (h *treatmentPlanCSVHandler) transformFTPsToSTPs(ctx context.Context, ftps []*ftp, threads int) (map[string][]byte, error) {
	errs := make(chan error, len(ftps))
	complete := make(chan *completedSTP, len(ftps))
	stps := make(map[string][]byte)
	done := 0
	started := 0
	for i := 0; i < threads && i < len(ftps); i++ {
		go h.transformFTPToSTP(ctx, *ftps[started], complete, errs)
		started++
	}
	for done != len(ftps) {
		select {
		case stp := <-complete:
			stps[stp.PathwayTag] = stp.STPJSON
			done++
		case err := <-errs:
			return nil, err
		}
		if started-done < threads && started < len(ftps) {
			go h.transformFTPToSTP(ctx, *ftps[started], complete, errs)
			started++
		}
	}

	return stps, nil
}

func (h *treatmentPlanCSVHandler) createGlobalFTPs(ftps []*ftp) error {
	// For now default this to be language_id EN
	dispenseUnitIDs, dispenseUnits, err := h.dataAPI.GetMedicationDispenseUnits(1)
	if err != nil {
		return err
	}
	dispenseUnitIDMapping := make(map[string]int64)
	for i, dispenseUnit := range dispenseUnits {
		dispenseUnitIDMapping[dispenseUnit] = dispenseUnitIDs[i]
	}
	ftpModels := make(map[int64][]*common.FavoriteTreatmentPlan)
	for _, ftp := range ftps {
		regimineSections := make([]*common.RegimenSection, 0, len(ftp.Sections))
		sectionKeys := make([]string, 0, len(ftp.Sections))
		for k := range ftp.Sections {
			sectionKeys = append(sectionKeys, k)
		}
		sort.Strings(sectionKeys)
		for _, k := range sectionKeys {
			steps := make([]*common.DoctorInstructionItem, 0, len(ftp.Sections[k].Steps))
			for _, st := range ftp.Sections[k].Steps {
				steps = append(steps, &common.DoctorInstructionItem{Text: st.Text})
			}
			regimineSection := &common.RegimenSection{
				Name:  ftp.Sections[k].Title,
				Steps: steps,
			}
			regimineSections = append(regimineSections, regimineSection)
		}
		regiminePlan := &common.RegimenPlan{
			Sections: regimineSections,
			Title:    ftp.SFTPName,
			Status:   "ACTIVE",
		}

		treatmentList := &common.TreatmentList{}
		rxKeys := make([]string, 0, len(ftp.Sections))
		for k := range ftp.RXs {
			rxKeys = append(rxKeys, k)
		}
		sort.Strings(rxKeys)
		for _, k := range rxKeys {
			msr, err := h.erxAPI.SelectMedication(0, ftp.RXs[k].Name, ftp.RXs[k].Dosage)
			if msr == nil || err != nil {
				var errMsg string
				if err != nil {
					errMsg = err.Error()
				}
				golog.Errorf("When ingesting FTP from CSV Dosespot failed to resolve medication by name and dosage of - %s, %s. %s", ftp.RXs[k].Name, ftp.RXs[k].Dosage, errMsg)
				return fmt.Errorf("When ingesting FTP from CSV Dosespot failed to resolve medication by name and dosage of - %s, %s. %s", ftp.RXs[k].Name, ftp.RXs[k].Dosage, errMsg)
			}
			treatment, _ := doctor_treatment_plan.CreateTreatmentFromMedication(msr, ftp.RXs[k].Dosage, ftp.RXs[k].Name)
			numberRefills := encoding.NullInt64{}
			numberRefills.Int64Value, err = strconv.ParseInt(ftp.RXs[k].Refills, 10, 64)
			if err != nil {
				return err
			}
			dispenseValue, err := strconv.ParseFloat(ftp.RXs[k].DispenseNumber, 64)
			if err != nil {
				return err
			}
			dispenseUnitID, ok := dispenseUnitIDMapping[ftp.RXs[k].DispenseType]
			if !ok {
				return fmt.Errorf("No dispense unit ID could be located for type %s", ftp.RXs[k].DispenseType)
			}
			treatment.NumberRefills = numberRefills
			treatment.DispenseValue = encoding.HighPrecisionFloat64(dispenseValue)
			treatment.DispenseUnitID = encoding.DeprecatedNewObjectID(dispenseUnitID)
			treatment.SubstitutionsAllowed = strings.ToLower(ftp.RXs[k].Substitutions) == "yes" || strings.ToLower(ftp.RXs[k].Substitutions) == "true"
			treatment.PatientInstructions = ftp.RXs[k].Sig
			treatmentList.Treatments = append(treatmentList.Treatments, treatment)
			treatmentList.Status = "ACTIVE"
		}

		scheduledMessages := make([]*common.TreatmentPlanScheduledMessage, 0, len(ftp.ScheduledMessages))
		scheduledMessagesKeys := make([]string, 0, len(ftp.Sections))
		for k := range ftp.ScheduledMessages {
			scheduledMessagesKeys = append(scheduledMessagesKeys, k)
		}
		sort.Strings(scheduledMessagesKeys)
		for _, k := range scheduledMessagesKeys {
			dur, err := ftp.ScheduledMessages[k].DurationInDays()
			if err != nil {
				return err
			}
			var attachments []*common.CaseMessageAttachment
			if ftp.ScheduledMessages[k].RequiresFollowup() {
				attachments = append(attachments, &common.CaseMessageAttachment{
					ItemType: "followup_visit",
					Title:    "Follow-Up Visit",
				})
			}
			scheduledMessages = append(scheduledMessages, &common.TreatmentPlanScheduledMessage{
				ScheduledDays: dur,
				Message:       ftp.ScheduledMessages[k].Message,
				Attachments:   attachments,
			})
		}

		resourceGuides := make([]*common.ResourceGuide, len(ftp.ResourceGuideTags))
		for i, rgt := range ftp.ResourceGuideTags {
			guide, err := h.dataAPI.GetResourceGuideFromTag(rgt)
			if err != nil {
				return err
			}
			resourceGuides[i] = guide
		}

		ftpModel := &common.FavoriteTreatmentPlan{
			Name:              ftp.SFTPName,
			Note:              ftp.Note.String(),
			RegimenPlan:       regiminePlan,
			TreatmentList:     treatmentList,
			ScheduledMessages: scheduledMessages,
			ResourceGuides:    resourceGuides,
			Lifecycle:         "ACTIVE",
		}

		pathway, err := h.dataAPI.PathwayForTag(ftp.FrameworkTag, api.PONone)
		if err != nil {
			return err
		}

		ftpModels[pathway.ID] = append(ftpModels[pathway.ID], ftpModel)
	}
	if err := h.dataAPI.InsertGlobalFTPsAndUpdateMemberships(ftpModels); err != nil {
		return err
	}
	return nil
}

func (h *treatmentPlanCSVHandler) transformFTPToSTP(ctx context.Context, ftp ftp, complete chan *completedSTP, errs chan error) {
	sftp := &treatment_plan.TreatmentPlanViewsResponse{
		HeaderViews: []views.View{
			treatment_plan.NewTPHeroHeaderView("Sample Treatment Plan", "Your doctor will personalize a treatment plan for you."),
		},
	}

	dummyRegimenPlan := &common.RegimenPlan{
		Sections: make([]*common.RegimenSection, len(ftp.Sections)),
	}

	instructionCiews := make([]views.View, len(ftp.Sections)+1)
	instructionCiews[0] = treatment_plan.NewTPCardView([]views.View{
		treatment_plan.NewTPTextView("title1_medium", "Your doctor will explain how to use your treatments together in a personalized care routine."),
	})
	for _, v := range sftp.HeaderViews {
		v.Validate("treatment")
	}

	sectionKeys := make([]string, 0, len(ftp.Sections))
	for k := range ftp.Sections {
		sectionKeys = append(sectionKeys, k)
	}
	sort.Strings(sectionKeys)
	sectionIndex := 1
	for i, k := range sectionKeys {
		sectionInstructionViews := make([]views.View, len(ftp.Sections[k].Steps)+1)
		sectionInstructionViews[0] = treatment_plan.NewTPCardTitleView(ftp.Sections[k].Title, "", false)
		dummyRegimenPlan.Sections[i] = &common.RegimenSection{
			Name:  ftp.Sections[k].Title,
			Steps: make([]*common.DoctorInstructionItem, len(ftp.Sections[k].Steps)),
		}
		for si, st := range ftp.Sections[k].Steps {
			sectionInstructionViews[si+1] = treatment_plan.NewTPListElement("bulleted", st.Text, si)
			dummyRegimenPlan.Sections[i].Steps[si] = &common.DoctorInstructionItem{
				Text: st.Text,
			}
		}
		instructionCiews[sectionIndex] = treatment_plan.NewTPCardView(sectionInstructionViews)
		sectionIndex++
	}
	sftp.InstructionViews = instructionCiews
	for _, v := range sftp.InstructionViews {
		v.Validate("treatment")
	}

	treatmentViews := make([]views.View, 0, len(ftp.RXs)+2)
	treatmentViews = append(treatmentViews, treatment_plan.NewTPCardView([]views.View{
		treatment_plan.NewTPTextView("title1_medium", "Your doctor will determine the right treatments for you."),
		treatment_plan.NewTPTextView("", "Prescriptions will be available to pick up at your preferred pharmacy."),
	}))
	treatmentList := &common.TreatmentList{}

	rxKeys := make([]string, 0, len(ftp.Sections))
	for k := range ftp.RXs {
		rxKeys = append(rxKeys, k)
	}
	sort.Strings(rxKeys)
	for _, k := range rxKeys {
		msr, err := h.erxAPI.SelectMedication(0, ftp.RXs[k].Name, ftp.RXs[k].Dosage)
		if err != nil {
			errs <- err
			return
		}
		if msr != nil {
			treatment, _ := doctor_treatment_plan.CreateTreatmentFromMedication(msr, ftp.RXs[k].Dosage, ftp.RXs[k].Name)
			treatment.PatientInstructions = ftp.RXs[k].Sig
			treatmentList.Treatments = append(treatmentList.Treatments, treatment)
			treatmentList.Status = "ACTIVE"
		}
	}
	if len(treatmentList.Treatments) != 0 {
		treatmentViews = append(treatmentViews, treatment_plan.GenerateViewsForTreatments(ctx, treatmentList, 0, h.dataAPI, false)...)
	}

	pd := &pharmacy.PharmacyData{
		AddressLine1: "1101 Market St",
		City:         "San Francisco",
		SourceID:     8561,
		Latitude:     37.77959,
		Longitude:    -122.41363,
		Name:         "Cvs/Pharmacy",
		Phone:        "4155581538",
		Source:       "surescripts",
		State:        "CA",
		URL:          "",
		Postal:       "94103",
	}

	treatmentViews = append(treatmentViews, treatment_plan.NewTPCardView([]views.View{
		treatment_plan.NewTPCardTitleView("Prescription Pickup", "", false),
		treatment_plan.NewPharmacyView("Your prescriptions should be ready soon. Call your pharmacy to confirm a pickup time.", nil, pd),
	}))
	sftp.TreatmentViews = treatmentViews
	for _, v := range sftp.TreatmentViews {
		v.Validate("treatment")
	}

	// create dummy treatment plan from which to generate views for single view treatment plan
	tp := &common.TreatmentPlan{
		TreatmentList: treatmentList,
		RegimenPlan:   dummyRegimenPlan,
		SentDate:      ptr.Time(time.Date(2015, 04, 15, 0, 0, 0, 0, time.UTC)),
	}

	contentViews := []views.View{
		treatment_plan.NewTPCardView(
			[]views.View{
				treatment_plan.NewTPTextView("title1_medium", "Your doctor will determine the right treatments for you and will explain how to use your treatments together in a personalized care routine."),
				treatment_plan.NewTPTextView("", "Prescriptions will be available to pick up at your preferred pharmacy."),
			}),
	}

	sftp.ContentViews = append(contentViews, treatment_plan.GenerateViewsForSingleViewTreatmentPlan(ctx, h.dataAPI, &treatment_plan.SingleViewTPConfig{
		TreatmentPlan:        tp,
		Pharmacy:             pd,
		ExcludeMessageButton: true,
		MessageText:          "You can always message your care team with any questions you have about your treatment plan.",
	})...)

	// validate content views
	for _, v := range sftp.ContentViews {
		if err := v.Validate("treatment"); err != nil {
			errs <- err
			return
		}
	}

	jsonData, err := json.Marshal(sftp)
	if err != nil {
		errs <- err
		return
	}

	if ftp.IsSTP {
		if err := h.dataAPI.CreatePathwaySTP(ftp.FrameworkTag, jsonData); err != nil {
			errs <- err
			return
		}
	}

	complete <- &completedSTP{
		PathwayTag: ftp.FrameworkTag,
		STPJSON:    jsonData,
	}
}

func parseFTPs(colData [][]string, threads int) ([]*ftp, error) {
	ftpCount := len(colData[0]) - 1
	iterationThreads := threads
	errs := make(chan error, ftpCount)
	complete := make(chan *ftp, ftpCount)
	ftps := make([]*ftp, ftpCount)

	// This is sub optimal threading here in the sense that any thread iteration group less than the number of FTPs is only a fast as the slowest thread
	for i := 0; i < ftpCount; i = i + iterationThreads {
		if ftpCount-i < threads {
			iterationThreads = ftpCount - i
		}
		for it := 0; it < iterationThreads; it++ {
			go parseFTP(colData, i+it+1, errs, complete)
		}
		completed := 0
		for completed != iterationThreads {
			select {
			case ftp, ok := <-complete:
				if !ok {
					return nil, errors.New("Something went wrong. Channel closed prematurely")
				}
				ftps[i+completed] = ftp
				completed++
			case err := <-errs:
				return nil, err
			}
		}
	}

	return ftps, nil
}

func parseFTP(colData [][]string, column int, errs chan error, complete chan *ftp) {
	ftp := newFTP()
	// This makes some assumptions that items that are numbered exist in series
	for i, data := range colData {
		t, d := data[0], data[column]
		t, d = strings.TrimSpace(t), strings.TrimSpace(d)
		switch {
		case "" == t:
		case "framework_name" == t:
			ftp.FrameworkName = d
		case "framework_tag" == t:
			ftp.FrameworkTag = d
		case "sftp_name" == t:
			ftp.SFTPName = d
		case "diagnosis" == t:
			ftp.Diagnosis = d
		case "note:welcome" == t:
			ftp.Note.Welcome = d
		case "note:condition_description" == t:
			ftp.Note.ConditionDescription = d
		case "note:md_recommendation" == t:
			ftp.Note.MDRecommendation = d
		case noteRXDescriptionRE.MatchString(t):
			if d != "" {
				ftp.Note.RXDescriptions[digitsRE.FindString(t)] = d
			}
		case noteAdditionalInfoRE.MatchString(t):
			if d != "" {
				ftp.Note.AdditionalInfo[digitsRE.FindString(t)] = d
			}
		case "note:closing" == t:
			ftp.Note.Closing = d
		case scheduledMessageDurationRE.MatchString(t):
			if d != "" {
				si := digitsRE.FindString(t)
				sm, ok := ftp.ScheduledMessages[si]
				if !ok {
					sm = scheduledMessage{}
				}
				sm.Duration = d
				ftp.ScheduledMessages[si] = sm
			}
		case scheduledMessageUnitRE.MatchString(t):
			if d != "" {
				si := digitsRE.FindString(t)
				sm, ok := ftp.ScheduledMessages[si]
				if !ok {
					sm = scheduledMessage{}
				}
				sm.Unit = d
				ftp.ScheduledMessages[si] = sm
			}
		case scheduledMessageAttachmentRE.MatchString(t):
			if d != "" {
				si := digitsRE.FindString(t)
				sm, ok := ftp.ScheduledMessages[si]
				if !ok {
					sm = scheduledMessage{}
				}
				sm.Attachment = d
				ftp.ScheduledMessages[si] = sm
			}
		case scheduledMessageRE.MatchString(t):
			if d != "" {
				si := digitsRE.FindString(t)
				sm, ok := ftp.ScheduledMessages[si]
				if !ok {
					sm = scheduledMessage{}
				}
				sm.Message = d
				ftp.ScheduledMessages[si] = sm
			}
		case rxDosageRE.MatchString(t):
			if d != "" {
				ri := digitsRE.FindString(t)
				r, ok := ftp.RXs[ri]
				if !ok {
					r = rx{}
				}
				r.Dosage = d
				ftp.RXs[ri] = r
			}
		case rxNameRE.MatchString(t):
			if d != "" {
				ri := digitsRE.FindString(t)
				r, ok := ftp.RXs[ri]
				if !ok {
					r = rx{}
				}
				r.Name = d
				ftp.RXs[ri] = r
			}
		case rxRefillsRE.MatchString(t):
			if d != "" {
				ri := digitsRE.FindString(t)
				r, ok := ftp.RXs[ri]
				if !ok {
					r = rx{}
				}
				r.Refills = d
				ftp.RXs[ri] = r
			}
		case rxDispenseNumberRE.MatchString(t):
			if d != "" {
				ri := digitsRE.FindString(t)
				r, ok := ftp.RXs[ri]
				if !ok {
					r = rx{}
				}
				r.DispenseNumber = d
				ftp.RXs[ri] = r
			}
		case rxDispenseTypeRE.MatchString(t):
			if d != "" {
				ri := digitsRE.FindString(t)
				r, ok := ftp.RXs[ri]
				if !ok {
					r = rx{}
				}
				r.DispenseType = d
				ftp.RXs[ri] = r
			}
		case rxSubstitutionsRE.MatchString(t):
			if d != "" {
				ri := digitsRE.FindString(t)
				r, ok := ftp.RXs[ri]
				if !ok {
					r = rx{}
				}
				r.Substitutions = d
				ftp.RXs[ri] = r
			}
		case rxSigRE.MatchString(t):
			if d != "" {
				ri := digitsRE.FindString(t)
				r, ok := ftp.RXs[ri]
				if !ok {
					r = rx{}
				}
				r.Sig = d
				ftp.RXs[ri] = r
			}
		case sectionTitleRE.MatchString(t):
			if d != "" {
				si := digitsRE.FindString(t)
				s, ok := ftp.Sections[si]
				if !ok {
					s = section{Steps: make([]step, 0)}
				}
				s.Title = d
				ftp.Sections[si] = s
			}
		case sectionStepRE.MatchString(t):
			if d != "" {
				si := digitsRE.FindAllString(t, 2)
				if len(si) != 2 {
					errs <- fmt.Errorf("Expected to find 2 digits in section step type `%s` but found %d", t, len(si))
					return
				}
				s, ok := ftp.Sections[si[0]]
				if !ok {
					s = section{}
				}
				order, err := strconv.ParseInt(si[1], 10, 64)
				if err != nil {
					errs <- fmt.Errorf("Expected to find 2 digits in section step type `%s` but found %d", t, len(si))
					return
				}
				s.Steps = append(s.Steps, step{Text: d, Order: order})
				ftp.Sections[si[0]] = s
			}
		case resourceGuideRE.MatchString(t):
			if d != "" {
				ftp.ResourceGuideTags = append(ftp.ResourceGuideTags, d)
			}
		case "sample_ftp" == t:
			ftp.IsSTP = (strings.ToLower(d) == "true" || strings.ToLower(d) == "yes")
		default:
			errs <- fmt.Errorf("Unable to identify row type '%s' in row %d", data[0], i)
			return
		}
	}
	complete <- ftp
}

func csvDataFromFile(f multipart.File) ([][]string, error) {
	reader := csv.NewReader(f)
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return data, nil
}
