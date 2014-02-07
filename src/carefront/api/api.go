package api

import (
	"carefront/libs/pharmacy"
	"errors"
	"net/http"
	"time"

	"carefront/common"
)

const (
	EN_LANGUAGE_ID              = 1
	DOCTOR_ROLE                 = "DOCTOR"
	PRIMARY_DOCTOR_STATUS       = "PRIMARY"
	PATIENT_ROLE                = "PATIENT"
	REVIEW_PURPOSE              = "REVIEW"
	CONDITION_INTAKE_PURPOSE    = "CONDITION_INTAKE"
	DIAGNOSE_PURPOSE            = "DIAGNOSE"
	FOLLOW_UP_WEEK              = "week"
	FOLLOW_UP_DAY               = "day"
	FOLLOW_UP_MONTH             = "month"
	CASE_STATUS_OPEN            = "OPEN"
	CASE_STATUS_SUBMITTED       = "SUBMITTED"
	CASE_STATUS_REVIEWING       = "REVIEWING"
	CASE_STATUS_CLOSED          = "CLOSED"
	CASE_STATUS_TRIAGED         = "TRIAGED"
	CASE_STATUS_TREATED         = "TREATED"
	CASE_STATUS_PHOTOS_REJECTED = "PHOTOS_REJECTED"
	HIPAA_AUTH                  = "hipaa"
	CONSENT_AUTH                = "consent"
	PATIENT_PHONE_HOME          = "Home"
	PATIENT_PHONE_WORK          = "Work"
	PATIENT_PHONE_CELL          = "Cell"
	ERX_STATUS_QUEUE            = "erx"
)

var (
	NoRowsError                  = errors.New("No rows exist")
	NoElligibileProviderInState  = errors.New("There are no providers elligible in the state the patient resides")
	NoRegimenPlanForPatientVisit = errors.New("There is no regimen plan for patient visit")
	NoDiagnosisResponseErr       = errors.New("No diagnosis response exists to the question queried tag queried with")
)

type PotentialAnswerInfo struct {
	PotentialAnswerId int64
	AnswerType        string
	Answer            string
	AnswerSummary     string
	AnswerTag         string
	Ordering          int64
}

type PatientAPI interface {
	GetPatientFromId(patientId int64) (patient *common.Patient, err error)
	GetPatientFromAccountId(accountId int64) (patient *common.Patient, err error)
	RegisterPatient(accountId int64, firstName, lastName, gender, zipCode, city, state, phone, phoneType string, dob time.Time) (*common.Patient, error)
	UpdatePatientWithERxPatientId(patientId, erxPatientId int64) error
	GetPatientIdFromAccountId(accountId int64) (int64, error)
	CreateCareTeamForPatient(patientId int64) (careTeam *common.PatientCareProviderGroup, err error)
	GetCareTeamForPatient(patientId int64) (careTeam *common.PatientCareProviderGroup, err error)
	CheckCareProvidingElligibility(shortState string, healthConditionId int64) (doctorId int64, err error)
	UpdatePatientAddress(patientId int64, addressLine1, addressLine2, city, state, zipCode, addressType string) error
	UpdatePatientPharmacy(patientId int64, pharmacyDetails *pharmacy.PharmacyData) error
	GetPatientPharmacySelection(patientId int64) (pharmacySelection *pharmacy.PharmacyData, err error)
	TrackPatientAgreements(patientId int64, agreements map[string]bool) error
	GetPatientFromPatientVisitId(patientVisitId int64) (patient *common.Patient, err error)
}

type PrescriptionStatus struct {
	TreatmentId        int64
	PrescriptionId     int64
	PrescriptionStatus string
	StatusTimeStamp    time.Time
}

type PatientVisitAPI interface {
	GetActivePatientVisitIdForHealthCondition(patientId, healthConditionId int64) (int64, error)
	GetLastCreatedPatientVisitIdForPatient(patientId int64) (int64, error)
	GetPatientIdFromPatientVisitId(patientVisitId int64) (int64, error)
	GetLatestSubmittedPatientVisit() (*common.PatientVisit, error)
	GetLatestClosedPatientVisitForPatient(patientId int64) (*common.PatientVisit, error)
	GetPatientVisitFromId(patientVisitId int64) (patientVisit *common.PatientVisit, err error)
	CreateNewPatientVisit(patientId, healthConditionId, layoutVersionId int64) (int64, error)
	UpdatePatientVisitStatus(patientVisitId int64, message, event string) error
	GetMessageForPatientVisitStatus(patientVisitId int64) (message string, err error)
	ClosePatientVisit(patientVisitId int64, event, message string) error
	SubmitPatientVisitWithId(patientVisitId int64) error
	UpdateFollowUpTimeForPatientVisit(patientVisitId, doctorId, currentTimeSinceEpoch, followUpValue int64, followUpUnit string) error
	GetFollowUpTimeForPatientVisit(patientVisitId int64) (followUp *common.FollowUp, err error)
	GetDiagnosisResponseToQuestionWithTag(questionTag string, doctorId, patientVisitId int64) ([]*common.AnswerIntake, error)
	AddDiagnosisSummaryForPatientVisit(summary string, patientVisitId, doctorId int64) error
	GetDiagnosisSummaryForPatientVisit(patientVisitId int64) (summary string, err error)
	DeactivatePreviousDiagnosisForPatientVisit(patientVisitId int64, doctorId int64) error
	RecordDoctorAssignmentToPatientVisit(patientVisitId, doctorId int64) error
	GetDoctorAssignedToPatientVisit(patientVisitId int64) (doctor *common.Doctor, err error)
	GetAdvicePointsForPatientVisit(patientVisitId int64) (advicePoints []*common.DoctorInstructionItem, err error)
	CreateAdviceForPatientVisit(advicePoints []*common.DoctorInstructionItem, patientVisitId int64) error
	CreateRegimenPlanForPatientVisit(regimenPlan *common.RegimenPlan) error
	GetRegimenPlanForPatientVisit(patientVisitId int64) (regimenPlan *common.RegimenPlan, err error)
	AddTreatmentsForPatientVisit(treatments []*common.Treatment, doctorId, patientVisitId int64) error
	GetTreatmentPlanForPatientVisit(patientVisitId int64) (treatmentPlan *common.TreatmentPlan, err error)
	GetTreatmentBasedOnPrescriptionId(erxId int64) (*common.Treatment, error)
	UpdateTreatmentsWithPrescriptionIds(treatments []*common.Treatment, doctorId, patientVisitId int64) error
	AddErxStatusEvent(treatments []*common.Treatment, statusEvent string) error
	GetPrescriptionStatusEventsForPatient(patientId int64) ([]*PrescriptionStatus, error)
}

type DoctorAPI interface {
	RegisterDoctor(accountId int64, firstName, lastName, gender string, dob time.Time) (int64, error)
	GetDoctorFromId(doctorId int64) (doctor *common.Doctor, err error)
	GetDoctorIdFromAccountId(accountId int64) (int64, error)
	GetRegimenStepsForDoctor(doctorId int64) (regimenSteps []*common.DoctorInstructionItem, err error)
	AddRegimenStepForDoctor(regimenStep *common.DoctorInstructionItem, doctorId int64) error
	UpdateRegimenStepForDoctor(regimenStep *common.DoctorInstructionItem, doctorId int64) error
	MarkRegimenStepToBeDeleted(regimenStep *common.DoctorInstructionItem, doctorId int64) error
	MarkRegimenStepsToBeDeleted(regimenSteps []*common.DoctorInstructionItem, doctorId int64) error
	GetAdvicePointsForDoctor(doctorId int64) (advicePoints []*common.DoctorInstructionItem, err error)
	AddOrUpdateAdvicePointForDoctor(advicePoint *common.DoctorInstructionItem, doctorId int64) error
	MarkAdvicePointToBeDeleted(advicePoint *common.DoctorInstructionItem, doctorId int64) error
	MarkAdvicePointsToBeDeleted(advicePoints []*common.DoctorInstructionItem, doctorId int64) error
	AssignPatientVisitToDoctor(doctorId, patientVisitId int64) error
	MarkPatientVisitAsOngoingInDoctorQueue(doctorId, patientVisitId int64) error
	UpdateStateForPatientVisitInDoctorQueue(doctorId, patientVisitId int64, currentState, updatedState string) error
	GetPendingItemsInDoctorQueue(doctorId int64) (doctorQueue []*DoctorQueueItem, err error)
	GetCompletedItemsInDoctorQueue(doctorId int64) (doctorQueue []*DoctorQueueItem, err error)
	GetMedicationDispenseUnits(languageId int64) (dispenseUnitIds []int64, dispenseUnits []string, err error)
	GetDrugInstructionsForDoctor(drugName, drugForm, drugRoute string, doctorId int64) (drugInstructions []*common.DoctorInstructionItem, err error)
	AddOrUpdateDrugInstructionForDoctor(drugName, drugForm, drugRoute string, drugInstructionToAdd *common.DoctorInstructionItem, doctorId int64) error
	DeleteDrugInstructionForDoctor(drugInstructionToDelete *common.DoctorInstructionItem, doctorId int64) error
	AddDrugInstructionsToTreatment(drugName, drugForm, drugRoute string, drugInstructions []*common.DoctorInstructionItem, treatmentId int64, doctorId int64) error
	AddFavoriteTreatment(treatment *common.DoctorFavoriteTreatment, doctorId int64) error
	GetFavoriteTreatments(doctorId int64) ([]*common.DoctorFavoriteTreatment, error)
	DeleteFavoriteTreatment(favoriteTreatment *common.DoctorFavoriteTreatment, doctorId int64) error
}

type IntakeAPI interface {
	GetPatientAnswersForQuestionsInGlobalSections(questionIds []int64, patientId int64) (map[int64][]*common.AnswerIntake, error)
	GetAnswersForQuestionsInPatientVisit(role string, questionIds []int64, roleId int64, patientVisitId int64) (map[int64][]*common.AnswerIntake, error)
	StoreAnswersForQuestion(role string, roleId, patientVisitId, layoutVersionId int64, answersToStorePerQuestion map[int64][]*common.AnswerIntake) error
	CreatePhotoAnswerForQuestionRecord(role string, roleId, questionId, patientVisitId, potentialAnswerId, layoutVersionId int64) (patientInfoIntakeId int64, err error)
	UpdatePhotoAnswerRecordWithObjectStorageId(patientInfoIntakeId, objectStorageId int64) error
	MakeCurrentPhotoAnswerInactive(role string, roleId, questionId, patientVisitId, potentialAnswerId, layoutVersionId int64) error
	RejectPatientVisitPhotos(patientVisitId int64) error
}

type IntakeLayoutAPI interface {
	GetQuestionType(questionId int64) (questionType string, err error)
	GetActiveLayoutInfoForHealthCondition(healthConditionTag, role, purpose string) (bucket, key, region string, err error)
	GetStorageInfoOfCurrentActivePatientLayout(languageId, healthConditionId int64) (bucket, key, region string, layoutVersionId int64, err error)
	GetStorageInfoOfCurrentActiveDoctorLayout(healthConditionId int64) (bucket, storage, region string, layoutVersionId int64, err error)
	GetStorageInfoOfActiveDoctorDiagnosisLayout(healthConditionId int64) (bucket, storage, region string, layoutVersionId int64, err error)
	GetLayoutVersionIdForPatientVisit(patientVisitId int64) (layoutVersionId int64, err error)
	GetStorageInfoForClientLayout(layoutVersionId, languageId int64) (bucket, key, region string, err error)
	MarkNewLayoutVersionAsCreating(objectId int64, syntaxVersion int64, healthConditionId int64, role, purpose, comment string) (int64, error)
	MarkNewPatientLayoutVersionAsCreating(objectId int64, languageId int64, layoutVersionId int64, healthConditionId int64) (int64, error)
	UpdatePatientActiveLayouts(layoutId int64, clientLayoutIds []int64, healthConditionId int64) error
	MarkNewDoctorLayoutAsCreating(objectId int64, layoutVersionId int64, healthConditionId int64) (int64, error)
	UpdateDoctorActiveLayouts(layoutId, doctorLayoutId, healthConditionId int64, purpose string) error
	GetGlobalSectionIds() ([]int64, error)
	GetSectionIdsForHealthCondition(healthConditionId int64) ([]int64, error)
	GetHealthConditionInfo(healthConditionTag string) (int64, error)
	GetSectionInfo(sectionTag string, languageId int64) (id int64, title string, err error)
	GetQuestionInfo(questionTag string, languageId int64) (*common.QuestionInfo, error)
	GetAnswerInfo(questionId int64, languageId int64) (answerInfos []PotentialAnswerInfo, err error)
	GetTipSectionInfo(tipSectionTag string, languageId int64) (id int64, tipSectionTitle string, tipSectionSubtext string, err error)
	GetTipInfo(tipTag string, languageId int64) (id int64, tip string, err error)
	GetSupportedLanguages() (languagesSupported []string, languagesSupportedIds []int64, err error)
}

type ObjectStorageDBAPI interface {
	CreateNewUploadCloudObjectRecord(bucket, key, region string) (int64, error)
	UpdateCloudObjectRecordToSayCompleted(id int64) error
}

type DataAPI interface {
	PatientAPI
	DoctorAPI
	PatientVisitAPI
	IntakeLayoutAPI
	ObjectStorageDBAPI
	IntakeAPI
}

type CloudStorageAPI interface {
	GetObjectAtLocation(bucket, key, region string) (rawData []byte, responseHeader http.Header, err error)
	GetSignedUrlForObjectAtLocation(bucket, key, region string, duration time.Time) (url string, err error)
	DeleteObjectAtLocation(bucket, key, region string) error
	PutObjectToLocation(bucket, key, region, contentType string, rawData []byte, duration time.Time, dataApi DataAPI) (int64, string, error)
}
