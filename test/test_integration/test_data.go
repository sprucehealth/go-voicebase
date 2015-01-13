package test_integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sprucehealth/backend/Godeps/_workspace/src/github.com/BurntSushi/toml"
	"github.com/sprucehealth/backend/Godeps/_workspace/src/github.com/samuel/go-metrics/metrics"
	"github.com/sprucehealth/backend/address"
	"github.com/sprucehealth/backend/analytics"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/apiservice/apipaths"
	"github.com/sprucehealth/backend/apiservice/router"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/common/config"
	"github.com/sprucehealth/backend/email"
	"github.com/sprucehealth/backend/environment"
	"github.com/sprucehealth/backend/libs/aws"
	"github.com/sprucehealth/backend/libs/aws/sns"
	"github.com/sprucehealth/backend/libs/aws/sqs"
	"github.com/sprucehealth/backend/libs/dispatch"
	"github.com/sprucehealth/backend/libs/erx"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/storage"
	"github.com/sprucehealth/backend/notify"
	"github.com/sprucehealth/backend/test"
)

func init() {
	apiservice.Testing = true
	dispatch.Testing = true
	golog.Default().SetLevel(golog.WARN)
}

type SMS struct {
	From, To, Text string
}

type SMSAPI struct {
	Sent []*SMS
	mu   sync.Mutex
}

func (s *SMSAPI) Send(from, to, text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Sent = append(s.Sent, &SMS{From: from, To: to, Text: text})
	return nil
}

func (s *SMSAPI) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.Sent)
}

type TestData struct {
	T                   *testing.T
	DataAPI             api.DataAPI
	AuthAPI             api.AuthAPI
	SMSAPI              *SMSAPI
	EmailService        *email.TestService
	ERxAPI              erx.ERxAPI
	DBConfig            config.DB
	Config              *router.Config
	CloudStorageService api.CloudStorageAPI
	DB                  *sql.DB
	AWSAuth             aws.Auth
	APIServer           *httptest.Server
}

func (d *TestData) AuthGet(url string, accountID int64) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("AccountID", strconv.FormatInt(accountID, 10))

	if accountID > 0 {
		token, err := d.AuthAPI.GetToken(accountID)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "token "+token)
	}

	return http.DefaultClient.Do(req)
}

func (d *TestData) AuthPost(url, bodyType string, body io.Reader, accountID int64) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", bodyType)
	return d.AuthPostWithRequest(req, accountID)
}

func (d *TestData) AuthPostJSON(url string, accountID int64, req, res interface{}) (*http.Response, error) {
	return d.authJSON("POST", url, accountID, req, res)
}

func (d *TestData) authJSON(method, url string, accountID int64, req, res interface{}) (*http.Response, error) {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(req); err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpRes, err := d.AuthPostWithRequest(httpReq, accountID)
	if err != nil {
		return httpRes, err
	}
	defer httpRes.Body.Close()
	return httpRes, json.NewDecoder(httpRes.Body).Decode(res)
}

func (d *TestData) AuthPostWithRequest(req *http.Request, accountID int64) (*http.Response, error) {
	if accountID > 0 {
		token, err := d.AuthAPI.GetToken(accountID)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "token "+token)
	}
	if req.Header.Get("S-Device-ID") == "" {
		req.Header.Set("S-Device-ID", "TEST")
	}
	return http.DefaultClient.Do(req)
}

func (d *TestData) AuthPut(url, bodyType string, body io.Reader, accountID int64) (*http.Response, error) {
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", bodyType)
	if accountID > 0 {
		token, err := d.AuthAPI.GetToken(accountID)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "token "+token)
	}

	return http.DefaultClient.Do(req)
}

func (d *TestData) AuthPutJSON(url string, accountID int64, req, res interface{}) (*http.Response, error) {
	return d.authJSON("PUT", url, accountID, req, res)
}

func (d *TestData) AuthDelete(url, bodyType string, body io.Reader, accountID int64) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", bodyType)
	if accountID > 0 {
		token, err := d.AuthAPI.GetToken(accountID)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "token "+token)
	}

	return http.DefaultClient.Do(req)
}

func (d *TestData) StartAPIServer(t *testing.T) {
	// close any previous api server
	if d.APIServer != nil {
		d.APIServer.Close()
	}

	// setup the restapi server
	mux := router.New(d.Config)
	d.APIServer = httptest.NewServer(mux)

	d.bootstrapData()
}

func (td *TestData) Close() {
	td.DB.Close()

	if td.APIServer != nil {
		td.APIServer.Close()
	}
	// put anything here that is global to the teardown process for integration tests
	teardownScript := os.Getenv(spruceProjectDirEnv) + "/src/github.com/sprucehealth/backend/test/test_integration/teardown_integration_test.sh"
	cmd := exec.Command(teardownScript, td.DBConfig.Name)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("CF_LOCAL_DB_INSTANCE=%s", td.DBConfig.Host),
		fmt.Sprintf("CF_LOCAL_DB_PORT=%d", td.DBConfig.Port),
		fmt.Sprintf("CF_LOCAL_DB_USERNAME=%s", td.DBConfig.User),
		fmt.Sprintf("CF_LOCAL_DB_PASSWORD=%s", td.DBConfig.Password),
	)
	err := cmd.Run()
	test.OK(td.T, err)
}

func setupTest() (*TestData, error) {
	testConf, err := getTestConf()
	if err != nil {
		return nil, err
	}
	dbConfig := testConf.DB

	if s := os.Getenv("CF_LOCAL_DB_INSTANCE"); s != "" {
		dbConfig.Host = s
	}
	if s := os.Getenv("CF_LOCAL_DB_PORT"); s != "" {
		dbConfig.Port, err = strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse CF_LOCAL_DB_PORT (%s) as int", s)
		}
	}
	if s := os.Getenv("CF_LOCAL_DB_USERNAME"); s != "" {
		dbConfig.User = s
		dbConfig.Password = os.Getenv("CF_LOCAL_DB_PASSWORD")
	}

	setupScript := os.Getenv(spruceProjectDirEnv) + "/src/github.com/sprucehealth/backend/test/test_integration/setup_integration_test.sh"
	cmd := exec.Command(setupScript)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("CF_LOCAL_DB_INSTANCE=%s", dbConfig.Host),
		fmt.Sprintf("CF_LOCAL_DB_PORT=%d", dbConfig.Port),
		fmt.Sprintf("CF_LOCAL_DB_USERNAME=%s", dbConfig.User),
		fmt.Sprintf("CF_LOCAL_DB_PASSWORD=%s", dbConfig.Password),
	)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	dbConfig.Name = strings.TrimSpace(out.String())
	db, err := dbConfig.ConnectMySQL(nil)
	if err != nil {
		return nil, err
	}

	conf := config.BaseConfig{}
	awsAuth, err := conf.AWSAuth()
	if err != nil {
		return nil, err
	}

	cloudStorageService := api.NewCloudStorageService(awsAuth)

	authTokenExpireDuration := time.Minute * 10
	authAPI, err := api.NewAuthAPI(db, authTokenExpireDuration, time.Minute*5, authTokenExpireDuration, time.Minute*5, nullHasher{})
	if err != nil {
		return nil, err
	}
	dataAPI, err := api.NewDataService(db, "api.spruce.local")
	if err != nil {
		return nil, err
	}

	testData := &TestData{
		DataAPI:             dataAPI,
		AuthAPI:             authAPI,
		DBConfig:            dbConfig,
		CloudStorageService: cloudStorageService,
		SMSAPI:              &SMSAPI{},
		EmailService:        &email.TestService{},
		DB:                  db,
		AWSAuth:             awsAuth,
		ERxAPI: erx.NewDoseSpotService(testConf.DoseSpot.ClinicID, testConf.DoseSpot.UserID,
			testConf.DoseSpot.ClinicKey, testConf.DoseSpot.SOAPEndpoint, testConf.DoseSpot.APIEndpoint, nil),
	}

	environment.SetCurrent("test")
	testData.Config = &router.Config{
		DataAPI:             testData.DataAPI,
		AuthAPI:             testData.AuthAPI,
		Dispatcher:          dispatch.New(),
		AuthTokenExpiration: authTokenExpireDuration,
		AnalyticsLogger:     analytics.DebugLogger{},
		AddressValidationAPI: &address.StubAddressValidationService{
			CityStateToReturn: &address.CityState{
				City:              "San Francisco",
				State:             "California",
				StateAbbreviation: "CA",
			},
		},
		PaymentAPI: &StripeStub{},
		NotifyConfigs: (*config.NotificationConfigs)(&map[string]*config.NotificationConfig{
			"iOS-Patient-Feature": &config.NotificationConfig{
				SNSApplicationEndpoint: "endpoint",
			},
		}),
		NotificationManager: notify.NewManager(testData.DataAPI, testData.AuthAPI, nil, testData.SMSAPI, &email.TestService{}, "", nil, metrics.NewRegistry()),
		ERxStatusQueue:      &common.SQSQueue{QueueService: &sqs.StubSQS{}, QueueURL: "local-status-erx"},
		ERxRoutingQueue:     &common.SQSQueue{QueueService: &sqs.StubSQS{}, QueueURL: "local-routing-erx"},
		ERxAPI: &erx.StubErxService{
			SelectMedicationFunc: func(clinicianID int64, name, strength string) (*erx.MedicationSelectResponse, error) {
				return &erx.MedicationSelectResponse{}, nil
			},
		},
		MedicalRecordQueue: &common.SQSQueue{QueueService: &sqs.StubSQS{}, QueueURL: "local-medrecord"},
		Stores: map[string]storage.Store{
			"media":          storage.NewS3(testData.AWSAuth, "us-east-1", "test-spruce-storage", "media"),
			"thumbnails":     storage.NewS3(testData.AWSAuth, "us-east-1", "test-spruce-storage", "thumbnails"),
			"medicalrecords": storage.NewTestStore(nil),
		},
		SNSClient:           &sns.MockSNS{PushEndpointToReturn: "push_endpoint"},
		MetricsRegistry:     metrics.NewRegistry(),
		CloudStorageAPI:     testData.CloudStorageService,
		DosespotConfig:      &config.DosespotConfig{},
		ERxRouting:          false,
		APIDomain:           "api.spruce.local",
		WebDomain:           "www.spruce.local",
		EmailService:        testData.EmailService,
		SMSAPI:              testData.SMSAPI,
		TwoFactorExpiration: 60,
	}

	return testData, nil
}

func (d *TestData) bootstrapData() {
	// FIX: We shouldn't have to signup this doctor, but currently
	// tests expect a default doctor to exist. Probably should get rid of this and update
	// tests to instantiate a doctor if one is needed
	SignupRandomTestDoctorInState("CA", d.T, d)

	// Upload first versions of the intake, review and diagnosis layouts
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	AddFileToMultipartWriter(writer, "intake", "intake-1-0-0.json", IntakeFileLocation, d.T)
	AddFileToMultipartWriter(writer, "review", "review-1-0-0.json", ReviewFileLocation, d.T)
	AddFileToMultipartWriter(writer, "diagnose", "diagnose-1-0-0.json", DiagnosisFileLocation, d.T)

	// specify the app versions and the platform information
	AddFieldToMultipartWriter(writer, "patient_app_version", "0.9.5", d.T)
	AddFieldToMultipartWriter(writer, "doctor_app_version", "1.2.3", d.T)
	AddFieldToMultipartWriter(writer, "platform", "iOS", d.T)

	err := writer.Close()
	test.OK(d.T, err)

	admin := CreateRandomAdmin(d.T, d)
	resp, err := d.AuthPost(d.APIServer.URL+apipaths.LayoutUploadURLPath, writer.FormDataContentType(), body, admin.AccountID.Int64())
	test.OK(d.T, err)
	defer resp.Body.Close()
	test.Equals(d.T, http.StatusOK, resp.StatusCode)

	// lets create the layout pair for followup visits
	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	AddFileToMultipartWriter(writer, "intake", "followup-intake-1-0-0.json", FollowupIntakeFileLocation, d.T)
	AddFileToMultipartWriter(writer, "review", "followup-review-1-0-0.json", FollowupReviewFileLocation, d.T)

	// specify the app versions and the platform information
	AddFieldToMultipartWriter(writer, "patient_app_version", "1.0.0", d.T)
	AddFieldToMultipartWriter(writer, "doctor_app_version", "1.0.0", d.T)
	AddFieldToMultipartWriter(writer, "platform", "iOS", d.T)

	err = writer.Close()
	test.OK(d.T, err)

	resp, err = d.AuthPost(d.APIServer.URL+apipaths.LayoutUploadURLPath, writer.FormDataContentType(), body, admin.AccountID.Int64())
	test.OK(d.T, err)
	defer resp.Body.Close()
	test.Equals(d.T, http.StatusOK, resp.StatusCode)

	// create drug descriptions for a handful of drugs
	// that we can easily reference when creating treatments in tests.
	// the reason we need to do this is because we use the drug description
	// in the database as the source of authority for treatments being added to TP or FTP.
	// In reality, the drug descriptions are added to the database (after being pulled down from the e-prescription service)
	// for each drug that the doctor selects on the app.
	for i := 0; i < 3; i++ {

		drugName := fmt.Sprintf("Drug%d", i+1)
		drugForm := fmt.Sprintf("Form%d", i+1)
		drugRoute := fmt.Sprintf("Route%d", i+1)
		drugStrength := fmt.Sprintf("Strength%d", i+1)

		err := d.DataAPI.SetDrugDescription(&api.DrugDescription{
			InternalName:   fmt.Sprintf("%s (%s - %s)", drugName, drugRoute, drugForm),
			DosageStrength: drugStrength,
			DrugDBIDs: map[string]string{
				erx.LexiDrugSynID:     "123",
				erx.LexiGenProductID:  "123",
				erx.LexiSynonymTypeID: "123",
				erx.NDC:               "1234",
			},
			OTC:             false,
			Schedule:        0,
			DrugName:        drugName,
			DrugForm:        drugForm,
			DrugRoute:       drugRoute,
			GenericDrugName: drugName,
		})
		test.OK(d.T, err)
	}

}

func SetupTest(t *testing.T) *TestData {
	CheckIfRunningLocally(t)
	t.Parallel()

	testData, err := setupTest()
	test.OK(t, err)

	return testData
}

func getTestConf() (*TestConf, error) {
	testConf := TestConf{}
	fileContents, err := ioutil.ReadFile(os.Getenv(spruceProjectDirEnv) + "/src/github.com/sprucehealth/backend/test/test.conf")
	if err != nil {
		return nil, err
	}
	_, err = toml.Decode(string(fileContents), &testConf)
	if err != nil {
		return nil, err
	}
	return &testConf, nil

}
