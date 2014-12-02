package common

import (
	"errors"
	"fmt"
	"time"

	"github.com/sprucehealth/backend/encoding"
	"github.com/sprucehealth/backend/pharmacy"
	"github.com/sprucehealth/backend/sku"
)

const (
	AttachmentTypeAudio         = "audio"
	AttachmentTypePhoto         = "photo"
	AttachmentTypeTreatmentPlan = "treatment_plan"
	AttachmentTypeVisit         = "visit"
)

const (
	ClaimerTypeConversationMessage = "conversation_message"
	ClaimerTypePhotoIntakeSection  = "patient_intake_photo_section"
)

type PhoneNumber struct {
	Phone    Phone  `json:"phone,omitempty"`
	Type     string `json:"phone_type,omitempty"`
	Status   string `json:"status"`
	Verified bool   `json:"verified"`
}

type Patient struct {
	PatientId         encoding.ObjectId      `json:"id,omitempty"`
	IsUnlinked        bool                   `json:"is_unlinked,omitempty"`
	FirstName         string                 `json:"first_name,omitempty"`
	LastName          string                 `json:"last_name,omiempty"`
	MiddleName        string                 `json:"middle_name,omitempty"`
	Suffix            string                 `json:"suffix,omitempty"`
	Prefix            string                 `json:"prefix,omitempty"`
	DOB               encoding.DOB           `json:"dob,omitempty"`
	Email             string                 `json:"email,omitempty"`
	Gender            string                 `json:"gender,omitempty"`
	ZipCode           string                 `json:"zip_code,omitempty"`
	CityFromZipCode   string                 `json:"-"`
	StateFromZipCode  string                 `json:"state_code,omitempty"`
	PhoneNumbers      []*PhoneNumber         `json:"phone_numbers,omitempty"`
	Status            string                 `json:"-"`
	AccountId         encoding.ObjectId      `json:"account_id,omitempty"`
	ERxPatientId      encoding.ObjectId      `json:"-"`
	PaymentCustomerId string                 `json:"-"`
	Pharmacy          *pharmacy.PharmacyData `json:"pharmacy,omitempty"`
	PatientAddress    *Address               `json:"address,omitempty"`
	PersonId          int64                  `json:"person_id"`
	PromptStatus      PushPromptStatus       `json:"prompt_status"`
	Training          bool                   `json:"is_training"`
}

type PCP struct {
	PatientID     int64  `json:"-"`
	PhysicianName string `json:"physician_full_name"`
	PhoneNumber   string `json:"phone_number"`
	PracticeName  string `json:"practice_name,omitempty"`
	Email         string `json:"email,omitempty"`
	FaxNumber     string `json:"fax_number,omitempty"`
}

func (p PCP) IsZero() bool {
	return p.PhysicianName == "" && p.PhoneNumber == "" && p.PracticeName == "" && p.Email == "" && p.FaxNumber == ""
}

type EmergencyContact struct {
	ID           int64  `json:"id,string"`
	PatientID    int64  `json:"-"`
	FullName     string `json:"full_name"`
	PhoneNumber  string `json:"phone_number"`
	Relationship string `json:"relationship"`
}

type Card struct {
	ID             encoding.ObjectId `json:"id,omitempty"`
	ThirdPartyID   string            `json:"third_party_id"`
	Fingerprint    string            `json:"fingerprint"`
	Token          string            `json:"token,omitempty"`
	Type           string            `json:"type"`
	ExpMonth       int64             `json:"exp_month"`
	ExpYear        int64             `json:"exp_year"`
	Last4          string            `json:"last4"`
	Label          string            `json:"label,omitempty"`
	BillingAddress *Address          `json:"address,omitempty"`
	IsDefault      bool              `json:"is_default,omitempty"`
	CreationDate   time.Time         `json:"creation_date"`
	ApplePay       bool              `json:"apple_pay"`
}

const (
	AlertSourcePatientVisitIntake = "PATIENT_VISIT_INTAKE"
	PAStatusActive                = "ACTIVE"
	PAStatusInactive              = "INACTIVE"
)

type Alert struct {
	Id           int64     `json:"-"`
	PatientId    int64     `json:"-"`
	Message      string    `json:"message"`
	CreationDate time.Time `json:"creation_date"`
	Source       string    `json:"-"`
	SourceId     int64     `json:"-"`
	Status       string    `json:"-"`
}

type Doctor struct {
	DoctorId            encoding.ObjectId `json:"id,omitempty"`
	FirstName           string            `json:"first_name,omitempty"`
	LastName            string            `json:"last_name,omitempty"`
	MiddleName          string            `json:"middle_name,omitempty"`
	Prefix              string            `json:"prefix,omitempty"`
	Suffix              string            `json:"suffix,omitempty"`
	ShortTitle          string            `json:"short_title,omitempty"`
	LongTitle           string            `json:"long_title,omitempty"`
	ShortDisplayName    string            `json:"short_display_name,omitempty"`
	LongDisplayName     string            `json:"long_display_name,omitempty"`
	DOB                 encoding.DOB      `json:"-"`
	Email               string            `json:"email"`
	Gender              string            `json:"-"`
	Status              string            `json:"-"`
	AccountId           encoding.ObjectId `json:"account_id"`
	CellPhone           Phone             `json:"phone"`
	LargeThumbnailID    string            `json:"-"`
	SmallThumbnailID    string            `json:"-"`
	LargeThumbnailURL   string            `json:"large_thumbnail_url,omitempty"`
	SmallThumbnailURL   string            `json:"small_thumbnail_url,omitempty"`
	DoseSpotClinicianId int64             `json:"-"`
	DoctorAddress       *Address          `json:"address,omitempty"`
	PersonId            int64             `json:"person_id"`
	PromptStatus        PushPromptStatus  `json:"prompt_status"`
	NPI                 string            `json:"npi,omitempty"`
	DEA                 string            `json:"dea,omitempty"`
	IsMA                bool              `json:"is_ma"`
}

const (
	PVStatusOpen      = "OPEN"
	PVStatusPending   = "PENDING"
	PVStatusSubmitted = "SUBMITTED"
	PVStatusReviewing = "REVIEWING"
	PVStatusTriaged   = "TRIAGED"
	PVStatusTreated   = "TREATED"
	PVStatusCharged   = "CHARGED"
	PVStatusRouted    = "ROUTED"
)

func NextPatientVisitStatus(currentStatus string) (string, error) {
	switch currentStatus {
	case PVStatusPending:
		return PVStatusOpen, nil
	case PVStatusOpen:
		return PVStatusSubmitted, nil
	case PVStatusSubmitted:
		return PVStatusReviewing, nil
	case PVStatusCharged:
		return PVStatusRouted, nil
	case PVStatusRouted:
		return PVStatusReviewing, nil
	case PVStatusReviewing:
		return "", fmt.Errorf("Ambiguous next step given it could be %s or %s", PVStatusTreated, PVStatusTriaged)
	case PVStatusTriaged, PVStatusTreated:
		return "", fmt.Errorf("No defined next step from %s", currentStatus)
	}

	return "", fmt.Errorf("Unknown current state: %s", currentStatus)
}

func SubmittedPatientVisitStates() []string {
	return []string{PVStatusSubmitted, PVStatusCharged, PVStatusRouted, PVStatusReviewing}
}

func TreatedPatientVisitStates() []string {
	return []string{PVStatusTreated, PVStatusTriaged}
}

func OpenPatientVisitStates() []string {
	return []string{PVStatusPending, PVStatusOpen}
}

type PatientVisit struct {
	PatientVisitId    encoding.ObjectId `json:"patient_visit_id,omitempty"`
	PatientCaseId     encoding.ObjectId `json:"case_id"`
	PatientId         encoding.ObjectId `json:"patient_id,omitempty"`
	CreationDate      time.Time         `json:"creation_date,omitempty"`
	SubmittedDate     time.Time         `json:"submitted_date,omitempty"`
	ClosedDate        time.Time         `json:"closed_date,omitempty"`
	HealthConditionId encoding.ObjectId `json:"health_condition_id,omitempty"`
	Status            string            `json:"status,omitempty"`
	LayoutVersionId   encoding.ObjectId `json:"layout_version_id,omitempty"`
	SKU               sku.SKU           `json:"-"`
}

type State struct {
	ID           int64
	Name         string
	Abbreviation string
	Country      string
}

type Address struct {
	Id           int64  `json:"-"`
	AddressLine1 string `json:"address_line_1"`
	AddressLine2 string `json:"address_line_2,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state"`
	ZipCode      string `json:"zip_code"`
	Country      string `json:"country"`
}

type CareProviderAssignment struct {
	ProviderRole      string     `json:"provider_role"`
	ProviderID        int64      `json:"provider_id"`
	FirstName         string     `json:"first_name,omitempty"`
	LastName          string     `json:"last_name,omitempty"`
	ShortTitle        string     `json:"short_title,omitempty"`
	LongTitle         string     `json:"long_title,omitempty"`
	ShortDisplayName  string     `json:"short_display_name,omitempty"`
	LongDisplayName   string     `json:"long_display_name,omitempty"`
	SmallThumbnailID  string     `json:"-"`
	LargeThumbnailID  string     `json:"-"`
	SmallThumbnailURL string     `json:"small_thumbnail_url,omitempty"`
	LargeThumbnailURL string     `json:"large_thumbnail_url,omitempty"`
	PatientId         int64      `json:"-"`
	HealthConditionId int64      `json:"-"`
	Status            string     `json:"-"`
	CreationDate      time.Time  `json:"assignment_date"`
	Expires           *time.Time `json:"-"`
}

type PatientCareTeam struct {
	Assignments []*CareProviderAssignment
}

type TreatmentPlanStatus string

var (
	TPStatusDraft     TreatmentPlanStatus = "DRAFT"
	TPStatusSubmitted TreatmentPlanStatus = "SUBMITTED"
	TPStatusActive    TreatmentPlanStatus = "ACTIVE"
	TPStatusInactive  TreatmentPlanStatus = "INACTIVE"
	TPStatusRXStarted TreatmentPlanStatus = "RX_STARTED"
)

func GetTreatmentPlanStatus(s string) (TreatmentPlanStatus, error) {
	switch t := TreatmentPlanStatus(s); t {
	case TPStatusDraft, TPStatusSubmitted, TPStatusActive, TPStatusInactive, TPStatusRXStarted:
		return t, nil
	}
	return TreatmentPlanStatus(""), fmt.Errorf("Unkown treatment plan status: %s", s)
}

func (t TreatmentPlanStatus) String() string {
	return string(t)
}

func (t *TreatmentPlanStatus) Scan(src interface{}) error {
	var err error
	switch ts := src.(type) {
	case string:
		*t, err = GetTreatmentPlanStatus(ts)
	case []byte:
		*t, err = GetTreatmentPlanStatus(string(ts))
	}
	return err
}

type FavoriteTreatmentPlan struct {
	Id            encoding.ObjectId `json:"id"`
	Name          string            `json:"name"`
	ModifiedDate  time.Time         `json:"modified_date,omitempty"`
	DoctorId      int64             `json:"-"`
	RegimenPlan   *RegimenPlan      `json:"regimen_plan,omitempty"`
	TreatmentList *TreatmentList    `json:"treatment_list,omitempty"`
	Note          string            `json:"note"`
}

func (f *FavoriteTreatmentPlan) EqualsTreatmentPlan(treatmentPlan *TreatmentPlan) bool {
	if f == nil || treatmentPlan == nil {
		return false
	}

	if !f.TreatmentList.Equals(treatmentPlan.TreatmentList) {
		return false
	}

	if !f.RegimenPlan.Equals(treatmentPlan.RegimenPlan) {
		return false
	}

	if f.Note != treatmentPlan.Note {
		return false
	}

	return true
}

func (f *FavoriteTreatmentPlan) Validate() error {
	if f == nil {
		return errors.New("Favorite treatment plan not provided")
	}

	if f.Name == "" {
		return errors.New("A favorite treatment plan requires a name")
	}

	// ensure that favorite treatment plan has atleast one of the sections filled out
	if (f.TreatmentList == nil || len(f.TreatmentList.Treatments) == 0) &&
		(f.RegimenPlan == nil || len(f.RegimenPlan.Sections) == 0) {
		return errors.New("A favorite treatment plan must have either a set of treatments or a regimen plan to be added")
	}

	return nil
}

type TreatmentPlan struct {
	Id            encoding.ObjectId           `json:"id,omitempty"`
	DoctorId      encoding.ObjectId           `json:"doctor_id,omitempty"`
	PatientCaseId encoding.ObjectId           `json:"case_id"`
	PatientId     int64                       `json:"patient_id,omitempty,string"`
	Status        TreatmentPlanStatus         `json:"status,omitempty"`
	CreationDate  time.Time                   `json:"creation_date"`
	SentDate      *time.Time                  `json:"sent_date,omitempty"`
	TreatmentList *TreatmentList              `json:"treatment_list"`
	RegimenPlan   *RegimenPlan                `json:"regimen_plan,omitempty"`
	Parent        *TreatmentPlanParent        `json:"parent,omitempty"`
	ContentSource *TreatmentPlanContentSource `json:"content_source,omitempty"`
	Note          string                      `json:"note,omitempty"`
}

func (d *TreatmentPlan) IsReadyForPatient() bool {
	switch d.Status {
	case TPStatusActive, TPStatusInactive:
		return true
	}

	return false
}

func (d *TreatmentPlan) IsActive() bool {
	switch d.Status {
	case TPStatusActive, TPStatusSubmitted, TPStatusRXStarted:
		return true
	}
	return false
}

func ActiveTreatmentPlanStates() []TreatmentPlanStatus {
	return []TreatmentPlanStatus{TPStatusActive, TPStatusSubmitted, TPStatusRXStarted}
}

func (d *TreatmentPlan) InDraftMode() bool {
	return d.Status == TPStatusDraft
}

const (
	TPParentTypeTreatmentPlan        = "TREATMENT_PLAN"
	TPParentTypePatientVisit         = "PATIENT_VISIT"
	TPContentSourceTypeFTP           = "FAVORITE_TREATMENT_PLAN"
	TPContentSourceTypeTreatmentPlan = "TREATMENT_PLAN"
)

// TreatmentPlanParent keeps track of the parent (either patient visit or previous treatment plan)
// so that we know how the treatment plan came into existence
type TreatmentPlanParent struct {
	ParentId     encoding.ObjectId `json:"parent_id"`
	ParentType   string            `json:"parent_type"`
	CreationDate time.Time         `json:"parent_creation_date"`
}

// TreatmentPlanContentSource keeps track of the source of the content
// for the treatment plan, given that doctor can start fresh with an empty treatment plan,
// from a previous treatment plan or from a favorite treatment plan when generating one for a patient
// Note that we indicate that the doctor started with an empty treatment plan by having nil for the
// content source in the treatment plan object.
// We also keep track of whether or not the treatment plan has deviated from the content source via the
// has_deviated flag
type TreatmentPlanContentSource struct {
	ID          encoding.ObjectId `json:"content_source_id"`
	Type        string            `json:"content_source_type"`
	HasDeviated bool              `json:"has_deviated"`
}

func (d *TreatmentPlan) Equals(other *TreatmentPlan) bool {
	if d == nil && other == nil {
		return true
	} else if d == nil || other == nil {
		return false
	}

	if !d.TreatmentList.Equals(other.TreatmentList) {
		return false
	}

	if !d.RegimenPlan.Equals(other.RegimenPlan) {
		return false
	}

	return true
}

type TreatmentList struct {
	Treatments []*Treatment `json:"treatments,omitempty"`
	Status     string       `json:"status,omitempty"`
}

func (t *TreatmentList) Equals(other *TreatmentList) bool {
	if t == nil || other == nil {
		return false
	}

	if len(t.Treatments) != len(other.Treatments) {
		return false
	}

	for i, treatment := range t.Treatments {
		if !treatment.Equals(other.Treatments[i]) {
			return false
		}
	}

	return true
}

type RefillRequestItem struct {
	Id                        int64             `json:"id,string"`
	RxRequestQueueItemId      int64             `json:"-"`
	ReferenceNumber           string            `json:"-"`
	PharmacyRxReferenceNumber string            `json:"-"`
	ApprovedRefillAmount      int64             `json:"approved_refill,string,omitempty"`
	ErxPatientId              int64             `json:"-"`
	PrescriptionId            int64             `json:"-"`
	PatientAddedForRequest    bool              `json:"-"`
	RequestDateStamp          time.Time         `json:"requested_date"`
	ClinicianId               int64             `json:"-"`
	Patient                   *Patient          `json:"patient,omitempty"`
	RequestedRefillAmount     string            `json:"requested_refill_amount,omitempty"`
	RequestedPrescription     *Treatment        `json:"requested_prescription,omitempty"`
	DispensedPrescription     *Treatment        `json:"dispensed_prescription"`
	Doctor                    *Doctor           `json:"-"`
	TreatmentPlanID           encoding.ObjectId `json:"treatment_plan_id,string,omitempty"`
	RxHistory                 []StatusEvent     `json:"refill_rx_history,omitempty"`
	Comments                  string            `json:"comments,omitempty"`
	DenialReason              string            `json:"denial_reason,omitempty"`
}

type DoctorTreatmentTemplate struct {
	Id        encoding.ObjectId `json:"id,omitempty"`
	Name      string            `json:"name"`
	Treatment *Treatment        `json:"treatment"`
	Status    string            `json:"-"`
}

const (
	STATE_ADDED    = "added"
	STATE_MODIFIED = "modified"
	STATE_DELETED  = "deleted"
)

type DoctorInstructionItem struct {
	ID       encoding.ObjectId `json:"id,omitempty"`
	ParentID encoding.ObjectId `json:"parent_id,omitempty"`
	Text     string            `json:"text"`
	Selected bool              `json:"selected,omitempty"`
	State    string            `json:"state,omitempty"`
	Status   string            `json:"-"`
}

func (d *DoctorInstructionItem) Equals(other *DoctorInstructionItem) bool {
	if d == nil || other == nil {
		return false
	}

	return d.Text == other.Text
}

type RegimenSection struct {
	ID    encoding.ObjectId        `json:"id,omitempty"`
	Name  string                   `json:"regimen_name"`
	Steps []*DoctorInstructionItem `json:"regimen_steps"`
}

type RegimenPlan struct {
	TreatmentPlanID encoding.ObjectId        `json:"treatment_plan_id,omitempty"`
	Sections        []*RegimenSection        `json:"regimen_sections"`
	AllSteps        []*DoctorInstructionItem `json:"all_regimen_steps,omitempty"`
	Title           string                   `json:"title,omitempty"`
	Status          string                   `json:"status,omitempty"`
}

func (r *RegimenPlan) Equals(other *RegimenPlan) bool {
	if r == nil && other == nil {
		return true
	} else if r == nil || other == nil {
		return false
	}

	// only compare regimen sections with atleast one step in them, because
	// the client currently sends regimen sections with no steps in them
	// making it harder to truly compare the contents of two regimen plans.
	rRegimenSections := getRegimenSectionsWithAtleastOneStep(r)
	otherRegimenSections := getRegimenSectionsWithAtleastOneStep(other)

	if len(rRegimenSections) != len(otherRegimenSections) {
		return false
	}

	// the ordering of the regimen sections and its steps have to be
	// exactly the same for the regimen plan to be considered equal
	for i, regimenSection := range rRegimenSections {
		if regimenSection.Name != otherRegimenSections[i].Name {
			return false
		}

		if len(regimenSection.Steps) != len(otherRegimenSections[i].Steps) {
			return false
		}

		for j, regimenStep := range regimenSection.Steps {
			if !regimenStep.Equals(otherRegimenSections[i].Steps[j]) {
				return false
			}
		}
	}

	return true
}

func getRegimenSectionsWithAtleastOneStep(r *RegimenPlan) []*RegimenSection {

	regimenSections := make([]*RegimenSection, 0, len(r.Sections))
	for _, regimenSection := range r.Sections {
		if len(regimenSection.Steps) > 0 {
			regimenSections = append(regimenSections, regimenSection)
		}
	}
	return regimenSections
}

type StatusEvent struct {
	ItemId            int64     `json:"-"`
	PrescriptionId    int64     `json:"-"`
	Status            string    `json:"status,omitempty"`
	InternalStatus    string    `json:"-"`
	StatusTimestamp   time.Time `json:"status_timestamp,omitempty"`
	ReportedTimestamp time.Time `json:"-"`
	StatusDetails     string    `json:"status_details,omitempty"`
}

type DrugDetails struct {
	Name              string
	NDC               string
	ImageURL          string
	OtherNames        string
	Description       string
	Route             string
	Tips              []string
	Warnings          []string
	CommonSideEffects []string
}

type Notification struct {
	Id              int64
	UID             string // Unique ID scoped to the patient.
	Timestamp       time.Time
	Expires         *time.Time
	Dismissible     bool
	DismissOnAction bool
	Priority        int
	Data            Typed
}

type HealthLogItem struct {
	Id        int64
	PatientId int64
	UID       string // Unique ID scoped to the patient.
	Timestamp time.Time
	Data      Typed
}

type Media struct {
	Id          int64
	Uploaded    time.Time
	UploaderID  int64
	URL         string
	Mimetype    string
	ClaimerType string
	ClaimerID   int64
}

type Person struct {
	Id       int64
	RoleType string
	RoleId   int64

	Patient *Patient
	Doctor  *Doctor
}

type CommunicationPreference struct {
	CommunicationType
	Id           int64
	AccountId    int64
	CreationDate time.Time
	Status       string
}

type SnoozeConfig struct {
	AccountID int64
	StartHour int
	NumHours  int
}

type PushConfigData struct {
	Id           int64
	AccountId    int64
	DeviceToken  string
	PushEndpoint string
	Platform
	PlatformVersion string
	AppType         string
	AppEnvironment  string
	AppVersion      string
	DeviceModel     string
	Device          string
	DeviceID        string
	CreationDate    time.Time
}

type ResourceGuideSection struct {
	ID      int64  `json:"id,string"`
	Ordinal int    `json:"ordinal"`
	Title   string `json:"title"`
}

type ResourceGuide struct {
	ID        int64       `json:"id,string"`
	SectionID int64       `json:"section_id,string"`
	Ordinal   int         `json:"ordinal"`
	Title     string      `json:"title"`
	PhotoURL  string      `json:"photo_url"`
	Layout    interface{} `json:"layout"`
}

type Account struct {
	ID               int64     `json:"id,string"`
	Role             string    `json:"role"`
	Email            string    `json:"email"`
	Registered       time.Time `json:"registered"`
	TwoFactorEnabled bool      `json:"two_factor_enabled"`
}

type AccountDevice struct {
	AccountID    int64     `json:"account_id,string"`
	DeviceID     string    `json:"device_id"`
	Verified     bool      `json:"verified"`
	VerifiedTime time.Time `json:"verified_time"`
	Created      time.Time `json:"created"`
}

type MedicalLicense struct {
	ID         int64                `json:"id,string"`
	DoctorID   int64                `json:"doctor_id,string"`
	State      string               `json:"state"`
	Number     string               `json:"number"`
	Expiration *time.Time           `json:"expiration,omitempty"`
	Status     MedicalLicenseStatus `json:"status"`
}

type BankAccount struct {
	ID                int64
	AccountID         int64
	StripeRecipientID string
	Created           time.Time
	Default           bool
	Verified          bool
	VerifyAmount1     int
	VerifyAmount2     int
	VerifyTransfer1ID string
	VerifyTransfer2ID string
	VerifyExpires     time.Time
}

type DoctorSearchResult struct {
	DoctorID  int64  `json:"doctor_id,string"`
	AccountID int64  `json:"account_id,string"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type CareProviderProfile struct {
	AccountID           int64     `json:"account_id,string"`
	FullName            string    `json:"full_name"`
	WhySpruce           string    `json:"why_spruce"`
	Qualifications      string    `json:"qualifications"`
	UndergraduateSchool string    `json:"undergraduate_school"`
	GraduateSchool      string    `json:"graduate_school"`
	MedicalSchool       string    `json:"medical_school"`
	Residency           string    `json:"residency"`
	Fellowship          string    `json:"fellowship"`
	Experience          string    `json:"experience"`
	Created             time.Time `json:"created"`
	Modified            time.Time `json:"modified"`
}

type MedicalRecord struct {
	ID         int64               `json:"id,string"`
	PatientID  int64               `json:"patient_id,string"`
	Status     MedicalRecordStatus `json:"status"`
	Error      string              `json:"error,omitempty"`
	StorageURL string              `json:"storage_url"`
	Requested  time.Time           `json:"requested"`
	Completed  *time.Time          `json:"completed,omitempty"`
}

type AnalyticsReport struct {
	ID             int64     `json:"id,string"`
	OwnerAccountID int64     `json:"owner_account_id,string"`
	Name           string    `json:"name"`
	Query          string    `json:"query"`
	Presentation   string    `json:"presentation"`
	Created        time.Time `json:"created"`
	Modified       time.Time `json:"modified"`
}

type AccountGroup struct {
	ID          int64    `json:"id,string"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions,omitempty"`
}
