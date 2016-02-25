package main

import (
	"database/sql"
	"io"
	"log"

	"github.com/cookieo9/resources-go"
	"github.com/rainycape/memcache"
	"github.com/samuel/go-librato/librato"
	"github.com/samuel/go-metrics/metrics"
	"github.com/sprucehealth/backend/analytics"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice/apipaths"
	"github.com/sprucehealth/backend/cmd/svc/restapi/mediastore"
	"github.com/sprucehealth/backend/diagnosis"
	"github.com/sprucehealth/backend/email"
	"github.com/sprucehealth/backend/events"
	"github.com/sprucehealth/backend/libs/branch"
	"github.com/sprucehealth/backend/libs/cfg"
	"github.com/sprucehealth/backend/libs/dispatch"
	"github.com/sprucehealth/backend/libs/erx"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/libs/ratelimit"
	"github.com/sprucehealth/backend/libs/sig"
	"github.com/sprucehealth/backend/libs/storage"
	"github.com/sprucehealth/backend/libs/stripe"
	"github.com/sprucehealth/backend/www"
	"github.com/sprucehealth/backend/www/router"
)

func buildWWW(
	conf *mainConfig,
	dataAPI api.DataAPI,
	applicationDB *sql.DB,
	authAPI api.AuthAPI,
	diagnosisAPI diagnosis.API,
	eventsClient events.Client,
	emailService email.Service,
	smsAPI api.SMSAPI,
	eRxAPI erx.ERxAPI,
	dispatcher *dispatch.Dispatcher,
	signer *sig.Signer,
	stores storage.StoreMap,
	rateLimiters ratelimit.KeyedRateLimiters,
	alog analytics.Logger,
	compressResponse bool,
	metricsRegistry metrics.Registry,
	cfgStore cfg.Store,
	memcacheClient *memcache.Client,
	proxiedSiteURL string,
) httputil.ContextHandler {
	stripeCli := &stripe.Client{
		SecretKey:      conf.Stripe.SecretKey,
		PublishableKey: conf.Stripe.PublishableKey,
	}

	var err error
	var analyticsDB *sql.DB
	if conf.AnalyticsDB.Host != "" {
		analyticsDB, err = conf.AnalyticsDB.ConnectPostgres()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		golog.Warningf("No analytics database configured")
	}

	var lc *librato.Client
	if conf.Stats.LibratoToken != "" && conf.Stats.LibratoUsername != "" {
		lc = &librato.Client{
			Username: conf.Stats.LibratoUsername,
			Token:    conf.Stats.LibratoToken,
		}
	}

	branchClient := branch.NewMemcachedClient(conf.BranchKey, memcacheClient)

	www.MustInitializeResources("resources")
	templateLoader := www.NewTemplateLoader(func(path string) (io.ReadCloser, error) {
		return resources.DefaultBundle.Open("templates/" + path)
	})

	return cfg.HTTPHandler(router.New(&router.Config{
		DataAPI:             dataAPI,
		AuthAPI:             authAPI,
		ApplicationDB:       applicationDB,
		DiagnosisAPI:        diagnosisAPI,
		SMSAPI:              smsAPI,
		ERxAPI:              eRxAPI,
		Dispatcher:          dispatcher,
		AnalyticsDB:         analyticsDB,
		AnalyticsLogger:     alog,
		FromNumber:          conf.Twilio.FromNumber,
		EmailService:        emailService,
		SupportEmail:        conf.Support.CustomerSupportEmail,
		WebDomain:           conf.WebDomain,
		APIDomain:           conf.APIDomain,
		StaticResourceURL:   conf.StaticResourceURL,
		StripeClient:        stripeCli,
		Signer:              signer,
		Stores:              stores,
		MediaStore:          mediastore.New("https://"+conf.APIDomain+apipaths.MediaURLPath, signer, stores.MustGet("media")),
		RateLimiters:        rateLimiters,
		WebPassword:         conf.WebPassword,
		TemplateLoader:      templateLoader,
		TwoFactorExpiration: conf.TwoFactorExpiration,
		ExperimentIDs:       conf.ExperimentID,
		LibratoClient:       lc,
		CompressResponse:    compressResponse,
		MetricsRegistry:     metricsRegistry.Scope("www"),
		EventsClient:        eventsClient,
		Cfg:                 cfgStore,
		BranchClient:        branchClient,
		ProxiedSiteURL:      proxiedSiteURL,
	}), cfgStore)
}
