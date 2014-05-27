package main

import (
	"carefront/address"
	"carefront/api"
	"carefront/apiservice"
	"carefront/app_worker"
	"carefront/common"
	"carefront/common/config"
	"carefront/demo"
	"carefront/doctor_queue"
	"carefront/homelog"
	"carefront/libs/aws"
	"carefront/libs/aws/sns"
	"carefront/libs/erx"
	"carefront/libs/golog"
	"carefront/libs/maps"
	"carefront/libs/payment/stripe"
	"carefront/libs/pharmacy"
	"carefront/libs/svcclient"
	"carefront/libs/svcreg"
	"carefront/messages"
	"carefront/notify"
	"carefront/patient"
	"carefront/patient_file"
	"carefront/patient_treatment_plan"
	"carefront/photos"
	"carefront/services/auth"
	thriftapi "carefront/thrift/api"
	"carefront/treatment_plan"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/SpruceHealth/go-proxy-protocol/proxyproto"
	"github.com/samuel/go-metrics/metrics"
	"github.com/subosito/twilio"
)

const (
	defaultMaxInMemoryPhotoMB = 2
)

func main() {
	conf := DefaultConfig
	_, err := config.Parse(&conf)
	if err != nil {
		log.Fatal(err)
	}

	if conf.Debug {
		golog.SetLevel(golog.DEBUG)
	} else if conf.Environment == "dev" {
		golog.SetLevel(golog.INFO)
	}

	conf.Validate()

	db, err := conf.DB.Connect(conf.BaseConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	metricsRegistry := metrics.NewRegistry()
	conf.StartReporters(metricsRegistry)

	if num, err := strconv.Atoi(config.MigrationNumber); err == nil {
		var latestMigration int
		if err := db.QueryRow("SELECT MAX(migration_id) FROM migrations").Scan(&latestMigration); err != nil {
			if !conf.Debug {
				log.Fatalf("Failed to query for latest migration: %s", err.Error())
			} else {
				log.Printf("Failed to query for latest migration: %s", err.Error())
			}
		}
		if latestMigration != num {
			if conf.Debug {
				golog.Warningf("Current database migration = %d, want %d", latestMigration, num)
			} else {
				// TODO: eventually make this Fatal once everything has been fully tested
				golog.Errorf("Current database migration = %d, want %d", latestMigration, num)
			}
		}
	} else if !conf.Debug {
		// TODO: eventually make this Fatal once everything has been fully tested
		golog.Errorf("MigrationNumber not set and not debug")
	}

	awsAuth, err := conf.AWSAuth()
	if err != nil {
		log.Fatalf("Failed to get AWS auth: %+v", err)
	}

	svcReg, err := conf.ServiceRegistry()
	if err != nil {
		log.Fatalf("Failed to create service registry: %+v", err)
	}

	var authApi thriftapi.Auth
	if conf.NoServices || conf.BaseConfig.ZookeeperHosts == "" {
		if conf.NoServices || conf.Debug {
			authApi = &auth.AuthService{
				DB:             db,
				ExpireDuration: time.Duration(conf.AuthTokenExpiration) * time.Second,
				RenewDuration:  time.Duration(conf.AuthTokenRenew) * time.Second,
				Hasher:         auth.NewBcryptHasher(0),
			}
		} else {
			log.Fatalf("No Zookeeper hosts defined and not running under debug")
		}
	} else {
		secureSvcClientBuilder, err := svcclient.NewThriftServiceClientBuilder(svcReg, svcreg.ServiceId{Environment: conf.Environment, Name: "secure"})
		if err != nil {
			log.Fatalf("Failed to create client builder for secure service: %+v", err)
		}
		secureSvcClient := svcclient.NewClient("restapi", 4, secureSvcClientBuilder, metricsRegistry.Scope("securesvc-client"))
		authApi = &thriftapi.AuthClient{Client: secureSvcClient}
	}

	if conf.InfoAddr != "" {
		go func() {
			log.Fatal(http.ListenAndServe(conf.InfoAddr, nil))
		}()
	}

	mapsService := maps.NewGoogleMapsService(metricsRegistry.Scope("google_maps_api"))
	doseSpotService := erx.NewDoseSpotService(conf.DoseSpot.ClinicId, conf.DoseSpot.UserId, conf.DoseSpot.ClinicKey, metricsRegistry.Scope("dosespot_api"))

	smartyStreetsService := &address.SmartyStreetsService{
		AuthId:    conf.SmartyStreets.AuthId,
		AuthToken: conf.SmartyStreets.AuthToken,
	}

	var erxStatusQueue *common.SQSQueue
	if conf.ERxQueue != "" {
		erxStatusQueue, err = common.NewQueue(awsAuth, aws.Regions[conf.AWSRegion], conf.ERxQueue)
		if err != nil {
			log.Fatal("Unable to get erx queue for sending prescriptions to: " + err.Error())
		}
	} else if conf.ERxRouting {
		log.Fatal("ERxQueue not configured but ERxRouting is enabled")
	}

	dataApi, err := api.NewDataService(db)
	if err != nil {
		log.Fatalf("Unable to initialize data service layer: %s", err)
	}

	snsClient := &sns.SNS{
		Region: aws.USEast,
		Client: &aws.Client{
			Auth: awsAuth,
		},
	}

	var twilioCli *twilio.Client
	if conf.Twilio != nil && conf.Twilio.AccountSid != "" && conf.Twilio.AuthToken != "" {
		twilioCli = twilio.NewClient(conf.Twilio.AccountSid, conf.Twilio.AuthToken, nil)
		notify.InitManager(dataApi, snsClient, twilioCli, conf.Twilio.FromNumber, conf.NotifiyConfigs, metricsRegistry.Scope("notify"))
	}

	homelog.InitListeners(dataApi)
	treatment_plan.InitListeners(dataApi)
	doctor_queue.InitListeners(dataApi)
	notify.InitListeners(dataApi)

	cloudStorageApi := api.NewCloudStorageService(awsAuth)
	photoAnswerCloudStorageApi := api.NewCloudStorageService(awsAuth)
	checkElligibilityHandler := &apiservice.CheckCareProvidingElligibilityHandler{DataApi: dataApi, AddressValidationApi: smartyStreetsService, StaticContentUrl: conf.StaticContentBaseUrl}
	updatePatientBillingAddress := &apiservice.UpdatePatientAddressHandler{DataApi: dataApi, AddressType: apiservice.BILLING_ADDRESS_TYPE}
	updatePatientPharmacyHandler := &apiservice.UpdatePatientPharmacyHandler{DataApi: dataApi, PharmacySearchService: pharmacy.GooglePlacesPharmacySearchService(0)}
	authenticateDoctorHandler := &apiservice.DoctorAuthenticationHandler{DataApi: dataApi, AuthApi: authApi}
	signupDoctorHandler := &apiservice.SignupDoctorHandler{DataApi: dataApi, AuthApi: authApi}
	patientTreatmentGuideHandler := patient_treatment_plan.NewPatientTreatmentGuideHandler(dataApi)
	doctorTreatmentGuideHandler := patient_treatment_plan.NewDoctorTreatmentGuideHandler(dataApi)
	patientVisitHandler := apiservice.NewPatientVisitHandler(dataApi, authApi, cloudStorageApi, photoAnswerCloudStorageApi)
	patientVisitReviewHandler := &patient_treatment_plan.PatientVisitReviewHandler{DataApi: dataApi}
	answerIntakeHandler := apiservice.NewAnswerIntakeHandler(dataApi)
	autocompleteHandler := &apiservice.AutocompleteHandler{DataApi: dataApi, ERxApi: doseSpotService, Role: api.PATIENT_ROLE}
	doctorTreatmentSuggestionHandler := &apiservice.AutocompleteHandler{DataApi: dataApi, ERxApi: doseSpotService, Role: api.DOCTOR_ROLE}
	doctorInstructionsHandler := apiservice.NewDoctorDrugInstructionsHandler(dataApi)
	doctorFollowupHandler := apiservice.NewPatientVisitFollowUpHandler(dataApi)
	doctorTreatmentTemplatesHandler := &apiservice.DoctorTreatmentTemplatesHandler{DataApi: dataApi}
	medicationStrengthSearchHandler := &apiservice.MedicationStrengthSearchHandler{DataApi: dataApi, ERxApi: doseSpotService}
	newTreatmentHandler := &apiservice.NewTreatmentHandler{DataApi: dataApi, ERxApi: doseSpotService}
	medicationDispenseUnitHandler := &apiservice.MedicationDispenseUnitsHandler{DataApi: dataApi}
	treatmentsHandler := &apiservice.TreatmentsHandler{
		DataApi: dataApi,
		ErxApi:  doseSpotService,
	}

	photoAnswerIntakeHandler := apiservice.NewPhotoAnswerIntakeHandler(dataApi, photoAnswerCloudStorageApi, conf.CaseBucket, conf.AWSRegion, conf.MaxInMemoryForPhotoMB*1024*1024)
	pharmacySearchHandler := &apiservice.PharmacyTextSearchHandler{PharmacySearchService: pharmacy.GooglePlacesPharmacySearchService(0), DataApi: dataApi, MapsService: mapsService}
	generateDoctorLayoutHandler := &apiservice.GenerateDoctorLayoutHandler{
		DataApi:                  dataApi,
		CloudStorageApi:          cloudStorageApi,
		DoctorLayoutBucket:       conf.DoctorLayoutBucket,
		DoctorVisualLayoutBucket: conf.DoctorVisualLayoutBucket,
		MaxInMemoryForPhoto:      conf.MaxInMemoryForPhotoMB,
		AWSRegion:                conf.AWSRegion,
		Purpose:                  api.REVIEW_PURPOSE,
	}
	generateDiagnoseLayoutHandler := &apiservice.GenerateDoctorLayoutHandler{
		DataApi:                  dataApi,
		CloudStorageApi:          cloudStorageApi,
		DoctorLayoutBucket:       conf.DoctorLayoutBucket,
		DoctorVisualLayoutBucket: conf.DoctorVisualLayoutBucket,
		MaxInMemoryForPhoto:      conf.MaxInMemoryForPhotoMB,
		AWSRegion:                conf.AWSRegion,
		Purpose:                  api.DIAGNOSE_PURPOSE,
	}
	pingHandler := apiservice.PingHandler(0)
	generateModelIntakeHandler := &apiservice.GenerateClientIntakeModelHandler{
		DataApi:             dataApi,
		CloudStorageApi:     cloudStorageApi,
		VisualLayoutBucket:  conf.VisualLayoutBucket,
		PatientLayoutBucket: conf.PatientLayoutBucket,
		AWSRegion:           conf.AWSRegion,
	}

	staticContentHandler := &apiservice.StaticContentHandler{
		DataApi:               dataApi,
		ContentStorageService: cloudStorageApi,
		BucketLocation:        conf.ContentBucket,
		Region:                conf.AWSRegion,
	}

	doctorPrescriptionErrorHandler := &apiservice.DoctorPrescriptionErrorHandler{
		DataApi: dataApi,
	}

	doctorPrescriptionErrorIgnoreHandler := &apiservice.DoctorPrescriptionErrorIgnoreHandler{
		DataApi: dataApi,
		ErxApi:  doseSpotService,
	}

	doctorRefillRequestHandler := &apiservice.DoctorRefillRequestHandler{
		DataApi:        dataApi,
		ErxApi:         doseSpotService,
		ErxStatusQueue: erxStatusQueue,
	}

	refillRequestDenialReasonsHandler := &apiservice.RefillRequestDenialReasonsHandler{
		DataApi: dataApi,
	}

	patientCardsHandler := &apiservice.PatientCardsHandler{
		DataApi:              dataApi,
		PaymentApi:           &stripe.StripeService{SecretKey: conf.StripeSecretKey},
		AddressValidationApi: smartyStreetsService,
	}

	doctorSubmitPatientVisitHandler := &apiservice.DoctorSubmitPatientVisitReviewHandler{DataApi: dataApi,
		ERxApi:         doseSpotService,
		ErxStatusQueue: erxStatusQueue,
		ERxRouting:     conf.ERxRouting}

	diagnosePatientHandler := &apiservice.DiagnosePatientHandler{
		DataApi:              dataApi,
		AuthApi:              authApi,
		LayoutStorageService: cloudStorageApi,
		Environment:          conf.Environment,
	}

	diagnosisSummaryHandler := &apiservice.DiagnosisSummaryHandler{DataApi: dataApi}
	doctorRegimenHandler := apiservice.NewDoctorRegimenHandler(dataApi)
	doctorAdviceHandler := apiservice.NewDoctorAdviceHandler(dataApi)

	doctorUpdatePatientPharmacyHandler := &apiservice.DoctorUpdatePatientPharmacyHandler{
		DataApi: dataApi,
	}

	doctorPharmacySearchHandler := &apiservice.DoctorPharmacySearchHandler{
		DataApi: dataApi,
		ErxApi:  doseSpotService,
	}
	doctorFavoriteTreatmentPlansHandler := &apiservice.DoctorFavoriteTreatmentPlansHandler{
		DataApi: dataApi,
	}
	doctorTreatmentPlanHandler := &apiservice.DoctorTreatmentPlanHandler{
		DataApi: dataApi,
	}

	mux := apiservice.NewAuthServeMux(authApi, metricsRegistry.Scope("restapi"))

	mux.Handle("/v1/content", staticContentHandler)
	mux.Handle("/v1/patient", patient.NewSignupHandler(dataApi, authApi))
	mux.Handle("/v1/patient/info", patient.NewUpdateHandler(dataApi))
	mux.Handle("/v1/patient/address/billing", updatePatientBillingAddress)
	mux.Handle("/v1/patient/pharmacy", updatePatientPharmacyHandler)
	mux.Handle("/v1/patient/treatment/guide", patientTreatmentGuideHandler)
	mux.Handle("/v1/patient/home", homelog.NewListHandler(dataApi))
	mux.Handle("/v1/patient/home/dismiss", homelog.NewDismissHandler(dataApi))
	mux.Handle("/v1/patient/isauthenticated", apiservice.NewIsAuthenticatedHandler(authApi))
	mux.Handle("/v1/patient/prompt_status", notify.NewPatientPromptStatusHandler(dataApi))
	mux.Handle("/v1/visit", patientVisitHandler)
	mux.Handle("/v1/visit/review", patientVisitReviewHandler)
	mux.Handle("/v1/check_eligibility", checkElligibilityHandler)
	mux.Handle("/v1/answer", answerIntakeHandler)
	mux.Handle("/v1/answer/photo", photoAnswerIntakeHandler)
	mux.Handle("/v1/authenticate", patient.NewAuthenticationHandler(dataApi, authApi, pharmacy.GooglePlacesPharmacySearchService(0), conf.StaticContentBaseUrl))
	mux.Handle("/v1/logout", patient.NewAuthenticationHandler(dataApi, authApi, pharmacy.GooglePlacesPharmacySearchService(0), conf.StaticContentBaseUrl))
	mux.Handle("/v1/ping", pingHandler)
	mux.Handle("/v1/autocomplete", autocompleteHandler)
	mux.Handle("/v1/pharmacy_search", pharmacySearchHandler)
	mux.Handle("/v1/doctor_layout", generateDoctorLayoutHandler)
	mux.Handle("/v1/diagnose_layout", generateDiagnoseLayoutHandler)
	mux.Handle("/v1/client_model", generateModelIntakeHandler)
	mux.Handle("/v1/credit_card", patientCardsHandler)
	mux.Handle("/v1/credit_card/default", patientCardsHandler)
	mux.Handle("/v1/notification/token", notify.NewNotificationHandler(dataApi, conf.NotifiyConfigs, snsClient))

	mux.Handle("/v1/photo", photos.NewHandler(dataApi, awsAuth, conf.PhotoBucket, conf.AWSRegion))
	mux.Handle("/v1/patient/conversation", messages.NewPatientConversationHandler(dataApi))
	mux.Handle("/v1/doctor/conversation", messages.NewDoctorConversationHandler(dataApi))
	mux.Handle("/v1/patient/conversation/messages", messages.NewPatientMessagesHandler(dataApi))
	mux.Handle("/v1/doctor/conversation/messages", messages.NewDoctorMessagesHandler(dataApi))
	mux.Handle("/v1/patient/conversation/read", messages.NewPatientReadHandler(dataApi))
	mux.Handle("/v1/doctor/conversation/read", messages.NewDoctorReadHandler(dataApi))
	mux.Handle("/v1/conversation/topics", messages.NewTopicsHandler(dataApi))

	mux.Handle("/v1/doctor/signup", signupDoctorHandler)
	mux.Handle("/v1/doctor/authenticate", authenticateDoctorHandler)
	mux.Handle("/v1/doctor/isauthenticated", apiservice.NewIsAuthenticatedHandler(authApi))
	mux.Handle("/v1/doctor/queue", doctor_queue.NewQueueHandler(dataApi))
	mux.Handle("/v1/doctor/treatment/templates", doctorTreatmentTemplatesHandler)

	mux.Handle("/v1/doctor/rx/error", doctorPrescriptionErrorHandler)
	mux.Handle("/v1/doctor/rx/error/resolve", doctorPrescriptionErrorIgnoreHandler)
	mux.Handle("/v1/doctor/rx/refill/request", doctorRefillRequestHandler)
	mux.Handle("/v1/doctor/rx/refill/denial_reasons", refillRequestDenialReasonsHandler)

	mux.Handle("/v1/doctor/patient/treatments", patient_file.NewDoctorPatientTreatmentsHandler(dataApi))
	mux.Handle("/v1/doctor/patient", patient_file.NewDoctorPatientHandler(dataApi, doseSpotService, smartyStreetsService))
	mux.Handle("/v1/doctor/patient/visits", patient_file.NewPatientVisitsHandler(dataApi))
	mux.Handle("/v1/doctor/patient/pharmacy", doctorUpdatePatientPharmacyHandler)
	mux.Handle("/v1/doctor/pharmacy", doctorPharmacySearchHandler)

	mux.Handle("/v1/doctor/visit/review", patient_file.NewDoctorPatientVisitReviewHandler(dataApi, pharmacy.GooglePlacesPharmacySearchService(0), cloudStorageApi, photoAnswerCloudStorageApi))
	mux.Handle("/v1/doctor/visit/treatment_plan", doctorTreatmentPlanHandler)
	mux.Handle("/v1/doctor/visit/diagnosis", diagnosePatientHandler)
	mux.Handle("/v1/doctor/visit/diagnosis/summary", diagnosisSummaryHandler)
	mux.Handle("/v1/doctor/visit/treatment/new", newTreatmentHandler)
	mux.Handle("/v1/doctor/visit/treatment/treatments", treatmentsHandler)
	mux.Handle("/v1/doctor/visit/treatment/medication_suggestions", doctorTreatmentSuggestionHandler)
	mux.Handle("/v1/doctor/visit/treatment/medication_strengths", medicationStrengthSearchHandler)
	mux.Handle("/v1/doctor/visit/treatment/medication_dispense_units", medicationDispenseUnitHandler)
	mux.Handle("/v1/doctor/visit/treatment/supplemental_instructions", doctorInstructionsHandler)
	mux.Handle("/v1/doctor/visit/treatment/guide", doctorTreatmentGuideHandler)
	mux.Handle("/v1/doctor/visit/regimen", doctorRegimenHandler)
	mux.Handle("/v1/doctor/visit/advice", doctorAdviceHandler)
	mux.Handle("/v1/doctor/visit/followup", doctorFollowupHandler)
	mux.Handle("/v1/doctor/visit/submit", doctorSubmitPatientVisitHandler)
	mux.Handle("/v1/doctor/favorite_treatment_plans", doctorFavoriteTreatmentPlansHandler)

	// add the api to create demo visits to every environment except production
	if conf.Environment != "prod" {
		mux.Handle("/v1/doctor/demo/patient_visit", demo.NewHandler(dataApi, cloudStorageApi, conf.AWSRegion, conf.Environment))
	}

	if conf.ERxRouting {
		app_worker.StartWorkerToUpdatePrescriptionStatusForPatient(dataApi, doseSpotService, erxStatusQueue, metricsRegistry.Scope("check_erx_status"))
		app_worker.StartWorkerToCheckForRefillRequests(dataApi, doseSpotService, metricsRegistry.Scope("check_rx_refill_requests"), conf.Environment)
	}

	s := &http.Server{
		Addr:           conf.ListenAddr,
		Handler:        mux,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if conf.TLSCert != "" && conf.TLSKey != "" {
		go func() {
			s.TLSConfig = &tls.Config{
				MinVersion:               tls.VersionTLS10,
				PreferServerCipherSuites: true,
				CipherSuites: []uint16{
					// Do not include RC4 or 3DES
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
					tls.TLS_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				},
			}
			if s.TLSConfig.NextProtos == nil {
				s.TLSConfig.NextProtos = []string{"http/1.1"}
			}

			cert, err := conf.ReadURI(conf.TLSCert)
			if err != nil {
				log.Fatal(err)
			}
			key, err := conf.ReadURI(conf.TLSKey)
			if err != nil {
				log.Fatal(err)
			}
			certs, err := tls.X509KeyPair(cert, key)
			if err != nil {
				log.Fatal(err)
			}

			s.TLSConfig.Certificates = []tls.Certificate{certs}

			conn, err := net.Listen("tcp", conf.TLSListenAddr)
			if err != nil {
				log.Fatal(err)
			}

			if conf.ProxyProtocol {
				conn = &proxyproto.Listener{Listener: conn}
			}

			ln := tls.NewListener(conn, s.TLSConfig)

			golog.Infof("Starting SSL server on %s...", conf.TLSListenAddr)
			log.Fatal(s.Serve(ln))
		}()
	}
	golog.Infof("Starting server on %s...", conf.ListenAddr)

	log.Fatal(s.ListenAndServe())
}
