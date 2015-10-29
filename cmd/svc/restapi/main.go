package main

import (
	"database/sql"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/sns"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/rainycape/memcache"
	"github.com/samuel/go-metrics/metrics"
	"github.com/sprucehealth/backend/analytics"
	"github.com/sprucehealth/backend/analytics/analisteners"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/boot"
	"github.com/sprucehealth/backend/common/config"
	"github.com/sprucehealth/backend/consul"
	"github.com/sprucehealth/backend/diagnosis"
	"github.com/sprucehealth/backend/email"
	"github.com/sprucehealth/backend/environment"
	"github.com/sprucehealth/backend/events"
	"github.com/sprucehealth/backend/libs/awsutil"
	"github.com/sprucehealth/backend/libs/cfg"
	"github.com/sprucehealth/backend/libs/dispatch"
	"github.com/sprucehealth/backend/libs/erx"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/libs/mandrill"
	"github.com/sprucehealth/backend/libs/mcutil"
	"github.com/sprucehealth/backend/libs/mux"
	"github.com/sprucehealth/backend/libs/ratelimit"
	"github.com/sprucehealth/backend/libs/sig"
	"github.com/sprucehealth/backend/libs/storage"
	"golang.org/x/net/context"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	defaultMaxInMemoryPhotoMB = 2
)

func connectDB(conf *mainConfig) *sql.DB {
	db, err := conf.DB.ConnectMySQL(conf.BaseConfig)
	if err != nil {
		log.Fatal(err)
	}

	if num, err := strconv.Atoi(boot.MigrationNumber); err == nil {
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

	return db
}

// gologStatsCollection implements the metrics.Collection interface and
// is used to export golog metrics.
type gologStatsCollection struct {
	stats golog.Stats
}

func (gsc *gologStatsCollection) Metrics() map[string]interface{} {
	golog.ReadStats(&gsc.stats)
	return map[string]interface{}{
		"critical": metrics.CounterValue(gsc.stats.Crit),
		"error":    metrics.CounterValue(gsc.stats.Err),
		"warning":  metrics.CounterValue(gsc.stats.Warn),
		"info":     metrics.CounterValue(gsc.stats.Info),
		"debug":    metrics.CounterValue(gsc.stats.Debug),
	}
}

func main() {
	conf := defaultConfig
	_, err := config.Parse(&conf)
	if err != nil {
		log.Fatal(err)
	}

	if conf.Debug {
		golog.Default().SetLevel(golog.DEBUG)
	} else if conf.Environment == "dev" {
		golog.Default().SetLevel(golog.INFO)
	}

	conf.Validate()

	dispatcher := dispatch.New()

	environment.SetCurrent(conf.Environment)

	stores := storage.StoreMap{}
	for name, c := range conf.Storage {
		switch strings.ToLower(c.Type) {
		default:
			log.Fatalf("Unknown storage type %s for name %s", c.Type, name)
		case "s3":
			s := storage.NewS3(conf.AWS(), c.Bucket, c.Prefix)
			stores[name] = s
		case "local":
			s, err := storage.NewLocalStore(c.Path)
			if err != nil {
				log.Fatalf("Failed to create store %s: %s", name, err)
			}
			stores[name] = s
		case "memory":
			stores[name] = storage.NewTestStore(nil)
		}
	}

	metricsRegistry := metrics.NewRegistry()

	var consulService *consul.Service
	var consulClient *consulapi.Client
	if conf.Consul.ConsulAddress != "" {
		consulClient, err = consulapi.NewClient(&consulapi.Config{
			Address:    conf.Consul.ConsulAddress,
			HttpClient: http.DefaultClient,
		})
		if err != nil {
			golog.Fatalf("Unable to instantiate new consul client: %s", err)
		}

		consulService, err = consul.RegisterService(consulClient, conf.Consul.ConsulServiceID, "restapi", nil, 0)
		if err != nil {
			log.Fatalf("Failed to register service with Consul: %s", err.Error())
		}
		if !consulService.WaitForRegistration(time.Second * 5) {
			// Only log an error here for now. Can change this behavior to something more appropriate later if necessary.
			golog.Errorf("Failed to acquire Consul session within 5 seconds")
		}
	} else {
		golog.Warningf("Consul address not specified")
	}

	var cfgStore cfg.Store
	if consulClient != nil {
		cfgStore, err = cfg.NewConsulStore(consulClient, "services/restapi/cfg", config.CfgDefs(), metricsRegistry.Scope("cfg"))
		if err != nil {
			golog.Fatalf("Failed to initialize consul cfg store: %s", err)
		}
	} else {
		cfgStore, err = cfg.NewLocalStore(config.CfgDefs())
		if err != nil {
			golog.Fatalf("Failed to initialize local cfg store: %s", err)
		}
	}

	defer func() {
		if consulService != nil {
			consulService.Deregister()
		}
	}()

	db := connectDB(&conf)
	defer db.Close()

	dataAPI, err := api.NewDataService(db, cfgStore, metricsRegistry.Scope("dataapi"))
	if err != nil {
		golog.Fatalf("Unable to initialize data service layer: %s", err)
	}

	if conf.InfoAddr != "" {
		http.Handle("/metrics", metrics.RegistryHandler(metricsRegistry))
		go func() {
			golog.Fatalf(http.ListenAndServe(conf.InfoAddr, nil).Error())
		}()
	}

	metricsRegistry.Add("log", &gologStatsCollection{})

	authAPI, err := api.NewAuthAPI(
		db,
		time.Duration(conf.RegularAuth.ExpireDuration)*time.Second,
		time.Duration(conf.RegularAuth.RenewDuration)*time.Second,
		time.Duration(conf.ExtendedAuth.ExpireDuration)*time.Second,
		time.Duration(conf.ExtendedAuth.RenewDuration)*time.Second,
		api.NewBcryptHasher(0),
	)
	if err != nil {
		golog.Fatalf(err.Error())
	}

	var smsAPI api.SMSAPI
	if twilioCli, err := conf.Twilio.Client(); err == nil {
		smsAPI = &twilioSMSAPI{twilioCli}
	} else if conf.Debug {
		golog.Infof(err.Error())
		smsAPI = loggingSMSAPI{}
	} else {
		golog.Fatalf(err.Error())
	}

	conf.StartReporters(metricsRegistry)

	sigKeys := make([][]byte, len(conf.SecretSignatureKeys))
	for i, k := range conf.SecretSignatureKeys {
		// No reason to decode the keys to binary. They'll be slightly longer
		// as ascii but include no less entropy.
		sigKeys[i] = []byte(k)
	}
	signer, err := sig.NewSigner(sigKeys, nil)
	if err != nil {
		log.Fatalf("Failed to create signer: %s", err.Error())
	}

	var alog analytics.Logger
	if conf.Analytics.LogPath != "" {
		var err error
		alog, err = analytics.NewFileLogger(conf.Analytics.LogPath, conf.Analytics.MaxEvents, time.Duration(conf.Analytics.MaxAge)*time.Second)
		if err != nil {
			log.Fatal(err)
		}
		if err := alog.Start(); err != nil {
			log.Fatal(err)
		}
	} else {
		if conf.Debug {
			alog = analytics.DebugLogger{}
		} else {
			alog = analytics.NullLogger{}
		}
	}

	// Deactivate the events system for now (but leave the code in place) while we reevaluate the data store and usefullness.
	eventsClient := events.NullClient{}

	analisteners.InitListeners(alog, dispatcher, eventsClient)

	snsCli := sns.New(conf.AWS())
	if conf.OfficeNotifySNSTopic != "" {
		initNotifyListener(dispatcher, snsCli, conf.OfficeNotifySNSTopic)
	}

	var memcacheCli *memcache.Client
	if conf.Memcached != nil {
		if m := conf.Memcached["cache"]; m != nil {
			var servers memcache.Servers
			if m.DiscoveryHost != "" {
				if m.DiscoveryInterval <= 0 {
					m.DiscoveryInterval = 60 * 5
				}
				d, err := awsutil.NewElastiCacheDiscoverer(m.DiscoveryHost, time.Second*time.Duration(m.DiscoveryInterval))
				if err != nil {
					log.Fatalf("Failed to discover memcached hosts: %s", err.Error())
				}
				servers = mcutil.NewElastiCacheServers(d)
			} else {
				servers = mcutil.NewHRWServer(m.Hosts)
			}
			memcacheCli = memcache.NewFromServers(servers)
		}
	}

	rateLimiters := ratelimit.KeyedRateLimiters(make(map[string]ratelimit.KeyedRateLimiter))
	if memcacheCli != nil {
		for n, c := range conf.RateLimiters {
			rateLimiters[n] = ratelimit.NewMemcache(memcacheCli, c.Max, c.Period)
		}
	}

	doseSpotService := erx.NewDoseSpotService(conf.DoseSpot.ClinicID, conf.DoseSpot.ProxyID,
		conf.DoseSpot.ClinicKey, conf.DoseSpot.SOAPEndpoint, conf.DoseSpot.APIEndpoint,
		metricsRegistry.Scope("dosespot_api"))

	diagnosisAPI, err := diagnosis.NewService(conf.DiagnosisDB)
	if err != nil {
		if conf.Debug {
			golog.Warningf("Failed to setup diagnosis service: %s", err)
		} else {
			golog.Fatalf("Failed to setup diagnosis service: %s", err)
		}
	}

	var emailService email.Service
	if conf.Mandrill.Key != "" {
		mand := mandrill.NewClient(conf.Mandrill.Key, conf.Mandrill.IPPool, metricsRegistry.Scope("email"))
		emailService = email.NewOptoutChecker(dataAPI, mand, cfgStore, dispatcher)
	} else if !environment.IsProd() && !environment.IsStaging() {
		emailService = email.NullService{}
	} else {
		golog.Fatalf("Mandrill not configured")
	}

	restAPIMux := buildRESTAPI(
		&conf, dataAPI, authAPI, diagnosisAPI, eventsClient, smsAPI, doseSpotService, memcacheCli,
		emailService, dispatcher, consulService, signer, stores, rateLimiters, alog, conf.CompressResponse,
		cfgStore, metricsRegistry, db)
	webMux := buildWWW(&conf, dataAPI, db, authAPI, diagnosisAPI, eventsClient, emailService, smsAPI,
		doseSpotService, dispatcher, signer, stores, rateLimiters, alog, conf.CompressResponse,
		metricsRegistry, cfgStore, memcacheCli)

	// Remove port numbers since the muxer doesn't include them in the match
	apiDomain := conf.APIDomain
	if i := strings.IndexByte(apiDomain, ':'); i > 0 {
		apiDomain = apiDomain[:i]
	}
	webDomain := conf.WebDomain
	if i := strings.IndexByte(webDomain, ':'); i > 0 {
		webDomain = webDomain[:i]
	}

	router := mux.NewRouter()
	router.Host(apiDomain).Handler(restAPIMux)
	router.Host(webDomain).Handler(webMux)

	// Redirect any unknown domains to the website. This will most likely be a
	// apex domain (e.g. sprucehealth.com -> www.sprucehealth.com).
	router.NotFoundHandler = httputil.ContextHandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if r.Host != apiDomain && r.Host != webDomain {
			// If apex domain (e.g. sprucehealth.com) then just rewrite host
			if idx := strings.IndexByte(r.Host, '.'); idx == strings.LastIndex(r.Host, ".") {
				u := *r.URL
				u.Scheme = "https"
				u.Host = webDomain
				http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
			} else {
				http.Redirect(w, r, "https://"+conf.WebDomain, http.StatusMovedPermanently)
			}
		} else {
			http.NotFound(w, r)
		}
		return
	})

	conf.SetupLogging()

	if conf.ErrorLogSNSTopic != "" {
		golog.Default().SetHandler(snsLogHandler(
			snsCli, conf.ErrorLogSNSTopic, strings.ToUpper(conf.Environment+"/"+conf.BaseConfig.AppName),
			golog.Default().Handler(), rateLimiters.Get("errorsns"), metricsRegistry.Scope("errorsns")))
	}

	serve(&conf, router)
}
