package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/rainycape/memcache"
	"github.com/samuel/go-metrics/metrics"
	"github.com/sprucehealth/backend/address"
	"github.com/sprucehealth/backend/analytics"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice/apipaths"
	"github.com/sprucehealth/backend/apiservice/router"
	"github.com/sprucehealth/backend/app_worker"
	"github.com/sprucehealth/backend/cmd/svc/restapi/mediastore"
	"github.com/sprucehealth/backend/cmd/svc/restapi/workers"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/consul"
	"github.com/sprucehealth/backend/cost"
	"github.com/sprucehealth/backend/demo"
	"github.com/sprucehealth/backend/diagnosis"
	"github.com/sprucehealth/backend/doctor_queue"
	"github.com/sprucehealth/backend/doctor_treatment_plan"
	"github.com/sprucehealth/backend/email"
	"github.com/sprucehealth/backend/email/campaigns"
	"github.com/sprucehealth/backend/environment"
	"github.com/sprucehealth/backend/events"
	"github.com/sprucehealth/backend/feedback"
	"github.com/sprucehealth/backend/libs/awsutil"
	"github.com/sprucehealth/backend/libs/cfg"
	"github.com/sprucehealth/backend/libs/conc"
	"github.com/sprucehealth/backend/libs/dispatch"
	"github.com/sprucehealth/backend/libs/erx"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/libs/ratelimit"
	"github.com/sprucehealth/backend/libs/sig"
	"github.com/sprucehealth/backend/libs/storage"
	"github.com/sprucehealth/backend/libs/stripe"
	"github.com/sprucehealth/backend/medrecord"
	"github.com/sprucehealth/backend/notify"
	"github.com/sprucehealth/backend/patient_case"
	"github.com/sprucehealth/backend/schedmsg"
	"github.com/sprucehealth/backend/surescripts/pharmacy"
	"golang.org/x/net/context"
)

func buildRESTAPI(
	conf *mainConfig,
	dataAPI api.DataAPI,
	authAPI api.AuthAPI,
	diagnosisAPI diagnosis.API,
	eventsClient events.Client,
	smsAPI api.SMSAPI,
	eRxAPI erx.ERxAPI,
	memcacheCli *memcache.Client,
	emailService email.Service,
	dispatcher *dispatch.Dispatcher,
	consulService *consul.Service,
	signer *sig.Signer,
	stores storage.StoreMap,
	rateLimiters ratelimit.KeyedRateLimiters,
	alog analytics.Logger,
	compressResponse bool,
	cfgStore cfg.Store,
	metricsRegistry metrics.Registry,
	applicationDB *sql.DB,
	errorNotifyHandler golog.Handler,
) httputil.ContextHandler {
	// Register the configs that will be used in different parts of the system
	registerCfgs(cfgStore)
	awsSession := conf.AWSSession()

	surescriptsPharmacySearch, err := pharmacy.NewSurescriptsPharmacySearch(conf.PharmacyDB)
	if err != nil {
		if conf.Debug {
			log.Printf("Unable to initialize pharmacy search: %s", err)
		} else {
			log.Fatalf("Unable to initialize pharmacy search: %s", err)
		}
	}

	var erxStatusQueue *common.SQSQueue
	if conf.ERxStatusQueue != "" {
		var err error
		erxStatusQueue, err = common.NewQueue(awsSession, conf.ERxStatusQueue)
		if err != nil {
			log.Fatalf("Unable to get erx queue for sending prescriptions to: %s", err.Error())
		}
	} else if conf.Debug {
		erxStatusQueue = &common.SQSQueue{
			QueueService: &awsutil.SQS{},
			QueueURL:     "ERxStatusQueue",
		}
	} else if conf.ERxRouting {
		log.Fatal("ERxStatusQueue not configured but ERxRouting is enabled")
	}

	var erxRoutingQueue *common.SQSQueue
	if conf.ERxRoutingQueue != "" {
		var err error
		erxRoutingQueue, err = common.NewQueue(awsSession, conf.ERxRoutingQueue)
		if err != nil {
			log.Fatalf("Unable to get erx queue for sending prescriptions to: %s", err.Error())
		}
	} else if conf.Debug {
		erxRoutingQueue = &common.SQSQueue{
			QueueService: &awsutil.SQS{},
			QueueURL:     "ERXRoutingQueue",
		}
	} else if conf.ERxRouting {
		log.Fatal("ERxRoutingQueue not configured but ERxRouting is enabled")
	}

	var medicalRecordQueue *common.SQSQueue
	if conf.MedicalRecordQueue != "" {
		medicalRecordQueue, err = common.NewQueue(awsSession, conf.MedicalRecordQueue)
		if err != nil {
			log.Fatalf("Failed to get queue for medical record requests: %s", err.Error())
		}
	} else if !conf.Debug {
		log.Fatal("MedicalRecordQueue not configured")
	} else {
		medicalRecordQueue = &common.SQSQueue{
			QueueService: &awsutil.SQS{},
			QueueURL:     "MedicalRecord",
		}
	}

	var visitQueue *common.SQSQueue
	if conf.VisitQueue != "" {
		visitQueue, err = common.NewQueue(awsSession, conf.VisitQueue)
		if err != nil {
			log.Fatalf("Failed to get queue for charging visits: %s", err.Error())
		}
	} else if !conf.Debug {
		log.Fatal("VisitQueue not configured")
	} else {
		visitQueue = &common.SQSQueue{
			QueueService: &awsutil.SQS{},
			QueueURL:     "Visit",
		}
	}

	snsClient := sns.New(awsSession)
	var addressValidationService address.Validator
	if conf.SmartyStreets == nil || !conf.SmartyStreets.IsSpecified() {
		if conf.Debug {
			addressValidationService = &localAddressValidationService{}
			golog.Warningf("Using stubbed address validation (which always returns San Francisco, CA for zipcode)")
		} else {
			golog.Fatalf("Smarty streets keys not specified")
		}
	} else {
		addressValidationService = &address.SmartyStreetsService{
			AuthID:    conf.SmartyStreets.AuthID,
			AuthToken: conf.SmartyStreets.AuthToken,
		}
	}

	notificationManager := notify.NewManager(dataAPI, authAPI, snsClient, smsAPI, emailService,
		conf.Twilio.FromNumber, conf.NotifiyConfigs, metricsRegistry.Scope("notify"))

	stripeClient := &stripe.Client{}
	if conf.TestStripe != nil && conf.TestStripe.SecretKey != "" {
		if conf.Environment == "prod" {
			golog.Warningf("Using test stripe key in production for patient")
		}
		stripeClient.SecretKey = conf.TestStripe.SecretKey
	} else {
		stripeClient.SecretKey = conf.Stripe.SecretKey
	}

	mediaStore := mediastore.New("https://"+conf.APIDomain+apipaths.MediaURLPath, signer, stores.MustGet("media"))

	var launchPromoStartDate *time.Time
	if conf.LaunchPromo != nil {
		launchPromoStartDate = &conf.LaunchPromo.StartDate
	}

	feedbackClient := feedback.NewDAL(applicationDB)

	_, muxHandler := router.New(&router.Config{
		DataAPI:                  dataAPI,
		AuthAPI:                  authAPI,
		Dispatcher:               dispatcher,
		AuthTokenExpiration:      time.Duration(conf.RegularAuth.ExpireDuration) * time.Second,
		MediaAccessExpiration:    10 * time.Minute,
		AddressValidator:         addressValidationService,
		PharmacySearchAPI:        surescriptsPharmacySearch,
		DiagnosisAPI:             diagnosisAPI,
		SNSClient:                snsClient,
		PaymentAPI:               stripeClient,
		MemcacheClient:           memcacheCli,
		NotifyConfigs:            conf.NotifiyConfigs,
		MinimumAppVersionConfigs: conf.MinimumAppVersionConfigs,
		DosespotConfig:           conf.DoseSpot,
		NotificationManager:      notificationManager,
		ERxRoutingQueue:          erxRoutingQueue,
		ERxStatusQueue:           erxStatusQueue,
		ERxAPI:                   eRxAPI,
		VisitQueue:               visitQueue,
		MedicalRecordQueue:       medicalRecordQueue,
		EmailService:             emailService,
		MetricsRegistry:          metricsRegistry,
		SMSAPI:                   smsAPI,
		Stores:                   stores,
		MediaStore:               mediaStore,
		RateLimiters:             rateLimiters,
		ERxRouting:               conf.ERxRouting,
		LaunchPromoStartDate:     launchPromoStartDate,
		NumDoctorSelection:       conf.NumDoctorSelection,
		JBCQMinutesThreshold:     conf.JBCQMinutesThreshold,
		CustomerSupportEmail:     conf.Support.CustomerSupportEmail,
		TechnicalSupportEmail:    conf.Support.TechnicalSupportEmail,
		APIDomain:                conf.APIDomain,
		WebDomain:                conf.WebDomain,
		APICDNDomain:             conf.APICDNDomain,
		StaticContentURL:         conf.StaticContentBaseURL,
		StaticResourceURL:        conf.StaticResourceURL,
		AWSRegion:                conf.AWSRegion,
		AnalyticsLogger:          alog,
		TwoFactorExpiration:      conf.TwoFactorExpiration,
		SMSFromNumber:            conf.Twilio.FromNumber,
		Cfg:                      cfgStore,
		ApplicationDB:            applicationDB,
		Signer:                   signer,
		FeedbackClient:           feedbackClient,
	})

	if !environment.IsProd() {
		demo.NewWorker(
			dataAPI,
			newConsulLock("service/restapi/training_cases", consulService, conf.Debug, errorNotifyHandler),
			conf.APIDomain,
			conf.AWSRegion,
		).Start()
	}

	notifyDoctorLock := newConsulLock("service/restapi/notify_doctor", consulService, conf.Debug, errorNotifyHandler)
	refillRequestCheckLock := newConsulLock("service/restapi/check_refill_request", consulService, conf.Debug, errorNotifyHandler)
	checkRxErrorsLock := newConsulLock("service/restapi/check_rx_error", consulService, conf.Debug, errorNotifyHandler)
	caseTimeoutLock := newConsulLock("service/restapi/case_timeout", consulService, conf.Debug, errorNotifyHandler)

	// Start worker to check for expired items in the global case queue
	doctor_queue.StartClaimedItemsExpirationChecker(dataAPI, alog, metricsRegistry.Scope("doctor_queue"))
	if conf.ERxRouting {
		app_worker.NewERxStatusWorker(
			dataAPI,
			eRxAPI,
			dispatcher,
			erxStatusQueue,
			metricsRegistry.Scope("check_erx_status"),
		).Start()
		app_worker.NewRefillRequestWorker(
			dataAPI,
			eRxAPI,
			refillRequestCheckLock,
			dispatcher,
			metricsRegistry.Scope("check_rx_refill_requests"),
		).Start()
		app_worker.NewERxErrorWorker(
			dataAPI,
			eRxAPI,
			dispatcher,
			checkRxErrorsLock,
			metricsRegistry.Scope("check_rx_errors"),
		).Start()
		doctor_treatment_plan.NewWorker(
			dataAPI, eRxAPI, dispatcher, erxRoutingQueue,
			erxStatusQueue, 0, metricsRegistry.Scope("erx_route"),
		).Start()
	}

	medrecord.NewWorker(
		dataAPI,
		diagnosisAPI,
		medicalRecordQueue,
		emailService,
		conf.Support.CustomerSupportEmail,
		conf.APIDomain,
		conf.WebDomain,
		signer,
		stores.MustGet("medicalrecords"),
		mediaStore,
		time.Duration(conf.RegularAuth.ExpireDuration)*time.Second,
		metricsRegistry.Scope("medrecord.worker"),
	).Start()

	schedmsg.StartWorker(dataAPI, authAPI, dispatcher, metricsRegistry.Scope("sched_msg"), 0)
	workers.StartAnalyticsWorker(dataAPI, metricsRegistry)

	cost.NewWorker(
		dataAPI,
		alog,
		dispatcher,
		stripeClient,
		emailService,
		visitQueue,
		metricsRegistry.Scope("visit_queue"),
		conf.VisitWorkerTimePeriodSeconds,
		conf.Support.CustomerSupportEmail,
		cfgStore,
	).Start()

	doctor_queue.NewWorker(
		dataAPI,
		authAPI,
		notifyDoctorLock,
		notificationManager,
		metricsRegistry.Scope("notify_doctors"),
	).Start()

	patient_case.NewWorker(
		dataAPI,
		caseTimeoutLock,
	).Start()

	campaigns.NewWorker(
		dataAPI,
		emailService,
		conf.WebDomain,
		signer,
		cfgStore,
		newConsulLock("service/restapi/email-campaigns", consulService, conf.Debug, errorNotifyHandler),
		metricsRegistry.Scope("email-campaigns-worker"),
	).Start()

	webRequestLogger := func(ctx context.Context, ev *httputil.RequestEvent) {
		av := &analytics.WebRequestEvent{
			Service:      "restapi",
			RequestID:    httputil.RequestID(ctx),
			Path:         ev.URL.Path,
			Timestamp:    analytics.Time(ev.Timestamp),
			StatusCode:   ev.StatusCode,
			Method:       ev.Request.Method,
			URL:          ev.URL.String(),
			RemoteAddr:   ev.RemoteAddr,
			ContentType:  ev.ResponseHeaders.Get("Content-Type"),
			UserAgent:    ev.Request.UserAgent(),
			Referrer:     ev.Request.Referer(),
			ResponseTime: int(ev.ResponseTime.Nanoseconds() / 1e3),
			Server:       ev.ServerHostname,
		}

		contextVals := []interface{}{
			"Method", av.Method,
			"URL", av.URL,
			"UserAgent", av.UserAgent,
			"RequestID", av.RequestID,
			"RemoteAddr", av.RemoteAddr,
			"StatusCode", av.StatusCode,
		}

		logMap, ok := ctx.Value(httputil.LogMapContextKey).(conc.Map)
		if ok {
			for k, v := range logMap.Snapshot() {
				contextVals = append(contextVals, k, v)
			}
		}

		log := golog.Context(contextVals...)

		if ev.Panic != nil {
			log.Criticalf("http: panic: %v\n%s", ev.Panic, ev.StackTrace)
		} else {
			log.Infof("apirequest")
		}
		dispatcher.PublishAsync(av)
	}

	h := httputil.SecurityHandler(muxHandler)
	h = httputil.LoggingHandler(h, webRequestLogger)
	h = httputil.MetricsHandler(h, metricsRegistry.Scope("restapi"))
	h = httputil.RequestIDHandler(h)
	h = httputil.DecompressRequest(h)
	if compressResponse {
		h = httputil.CompressResponse(h)
	}
	return h
}

func registerCfgs(cfgStore cfg.Store) {
	cfgStore.Register(cost.GlobalFirstVisitFreeEnabled)

}
