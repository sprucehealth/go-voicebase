package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/sprucehealth/backend/common/config"
	"github.com/sprucehealth/backend/email"
	"github.com/sprucehealth/backend/surescripts/pharmacy"
	"github.com/sprucehealth/backend/third_party/github.com/subosito/twilio"
)

type TwilioConfig struct {
	AccountSid string `long:"twilio_account_sid" description:"Twilio AccountSid"`
	AuthToken  string `long:"twilio_auth_token" description:"Twilio AuthToken"`
	FromNumber string `long:"twilio_from_number" description:"Twilio From Number for Messages"`

	client *twilio.Client
}

func (c *TwilioConfig) Client() (*twilio.Client, error) {
	if c.client != nil {
		return c.client, nil
	}
	if c == nil {
		return nil, fmt.Errorf("Twilio config does not exist")
	}
	if c.AccountSid == "" {
		return nil, fmt.Errorf("Twilio.AccountSid not set")
	}
	if c.AuthToken == "" {
		return nil, fmt.Errorf("Twilio.AuthToken not set")
	}
	c.client = twilio.NewClient(c.AccountSid, c.AuthToken, nil)
	return c.client, nil
}

type StripeConfig struct {
	SecretKey      string `long:"secret_key" description:"Secrey Key for stripe"`
	PublishableKey string `long:"publishable_key" description:"Publishable Key for stripe"`
}

type SmartyStreetsConfig struct {
	AuthId    string `long:"auth_id" description:"Auth id for smarty streets"`
	AuthToken string `long:"auth_token" description:"Auth token for smarty streets"`
}

type AnalyticsConfig struct {
	LogPath   string `long:"analytics_log_path" description:"Path to store analytics logs"`
	MaxEvents int    `long:"analytics_max_events" description:"Max number of events per log file before rotating"`
	MaxAge    int    `long:"analytics_max_age" description:"Max age of a log file in seconds before rotating"`
}

type SupportConfig struct {
	TechnicalSupportEmail string `long:"technical_support_email" description:"Email address for technical support"`
	CustomerSupportEmail  string `long:"customer_support_email" description:"Customer support email address"`
}

type StorageConfig struct {
	Type string
	// S3
	Region string
	Bucket string
	Prefix string
}

type AuthTokenConfig struct {
	ExpireDuration int `long:"auth_token_expire" description:"Expiration time in seconds for the auth token"`
	RenewDuration  int `long:"auth_token_renew" description:"Time left below which to renew the auth token"`
}

type ConsulConfig struct {
	ConsulAddress   string `long:"consul_address" description:"Consul HTTP API host:port"`
	ConsulServiceID string `long:"consul_service_id" description:"Service ID for Consul. Only needed when running more than one instance on a host."`
}

type Config struct {
	*config.BaseConfig
	ProxyProtocol                bool                             `long:"proxy_protocol" description:"Enable if behind a proxy that uses the PROXY protocol"`
	ListenAddr                   string                           `short:"l" long:"listen" description:"Address and port on which to listen (e.g. 127.0.0.1:8080)"`
	TLSListenAddr                string                           `long:"tls_listen" description:"Address and port on which to listen (e.g. 127.0.0.1:8080)"`
	TLSCert                      string                           `long:"tls_cert" description:"Path of SSL certificate"`
	TLSKey                       string                           `long:"tls_key" description:"Path of SSL private key"`
	APIDomain                    string                           `long:"api_domain" description:"Domain of REST API"`
	WebDomain                    string                           `long:"www_domain" description:"Domain of website"`
	InfoAddr                     string                           `long:"info_addr" description:"Address to listen on for the info server"`
	DB                           *config.DB                       `group:"Database" toml:"database"`
	AnalyticsDB                  *config.DB                       `group:"AnalyticsDatabase" toml:"AnalyticsDatabase"`
	MaxInMemoryForPhotoMB        int64                            `long:"max_in_memory_photo" description:"Amount of data in MB to be held in memory when parsing multipart form data"`
	ContentBucket                string                           `long:"content_bucket" description:"S3 Bucket name for all static content"`
	CaseBucket                   string                           `long:"case_bucket" description:"S3 Bucket name for case information"`
	Debug                        bool                             `long:"debug" description:"Enable debugging"`
	DoseSpotUserId               string                           `long:"dose_spot_user_id" description:"DoseSpot UserId for eRx integration"`
	NoServices                   bool                             `long:"noservices" description:"Disable connecting to remote services"`
	ERxRouting                   bool                             `long:"erx_routing" description:"Disable sending of prescriptions electronically"`
	ERxRoutingQueue              string                           `long:"erx_routing_queue" description:"ERx Routing Queue"`
	ERxStatusQueue               string                           `long:"erx_status_queue" description:"Erx queue name"`
	MedicalRecordQueue           string                           `long:"medical_record_queue" description:"Queue name for background generation of medical record"`
	VisitQueue                   string                           `long:"visit_queue" description:"Queue name for background charging and routing of patient visits"`
	VisitWorkerTimePeriodSeconds int                              `long:"visit_worker_time_period" description:"Time period between worker checking for messages in queue"`
	JBCQMinutesThreshold         int                              `long:"jbcq_minutes_threshold" description:"Threshold of inactivity between activities"`
	OnboardingURLExpires         int64                            `long:"onboarding_url_expire_duration" description:"duration for which an onboarding url will stay valid"`
	RegularAuth                  *AuthTokenConfig                 `group:"regular_auth" toml:"regular_auth"`
	ExtendedAuth                 *AuthTokenConfig                 `group:"extended_auth" toml:"extended_auth"`
	StaticContentBaseUrl         string                           `long:"static_content_base_url" description:"URL from which to serve static content"`
	Twilio                       *TwilioConfig                    `group:"Twilio" toml:"twilio"`
	DoseSpot                     *config.DosespotConfig           `group:"Dosespot" toml:"dosespot"`
	Consul                       *ConsulConfig                    `group:"Consul" toml:"consul"`
	SmartyStreets                *SmartyStreetsConfig             `group:"smarty_streets" toml:"smarty_streets"`
	TestStripe                   *StripeConfig                    `group:"test_stripe" toml:"test_stripe"`
	Stripe                       *StripeConfig                    `group:"stripe" toml:"stripe"`
	MinimumAppVersionConfigs     *config.MinimumAppVersionConfigs `group:"minimum_app_version"  toml:"minimum_app_version"`
	IOSDeeplinkScheme            string                           `long:"ios_deeplink_scheme" description:"Scheme for iOS deep-links (e.g. spruce://)"`
	NotifiyConfigs               *config.NotificationConfigs      `group:"notification" toml:"notification"`
	Analytics                    *AnalyticsConfig                 `group:"Analytics" toml:"analytics"`
	Support                      *SupportConfig                   `group:"support" toml:"support"`
	Email                        *email.Config                    `group:"email" toml:"email"`
	PharmacyDB                   *pharmacy.Config                 `group:"pharmacy_database" toml:"pharmacy_database"`
	Storage                      map[string]*StorageConfig        `group:"storage" toml:"storage"`
	StaticResourceURL            string                           `long:"static_url" description:"URL prefix for static resources"`
	WebPassword                  string                           `long:"web_password" description:"Password to access website"`
	TwoFactorExpiration          int                              `description:"Time to live of two factor auth token in seconds"`
	OfficeNotifySNSTopic         string                           `description:"SNS Topic to send submitted visit notifications"`
	// Secret keys used for generating signatures
	SecretSignatureKeys []string
}

var DefaultConfig = Config{
	BaseConfig: &config.BaseConfig{
		AppName: "restapi",
	},
	DB: &config.DB{
		Name: "carefront",
		Host: "127.0.0.1",
		Port: 3306,
	},
	Twilio:                &TwilioConfig{},
	APIDomain:             "api.sprucehealth.com",
	WebDomain:             "www.sprucehealth.com",
	ListenAddr:            ":8080",
	TLSListenAddr:         ":8443",
	InfoAddr:              ":9000",
	CaseBucket:            "carefront-cases",
	MaxInMemoryForPhotoMB: defaultMaxInMemoryPhotoMB,
	RegularAuth: &AuthTokenConfig{
		ExpireDuration: 60 * 60 * 24 * 2,
		RenewDuration:  60 * 60 * 36,
	},
	ExtendedAuth: &AuthTokenConfig{
		ExpireDuration: 60 * 60 * 24 * 30 * 2,
		RenewDuration:  60 * 60 * 24 * 45,
	},
	OnboardingURLExpires: 60 * 60 * 24 * 14,
	IOSDeeplinkScheme:    "spruce",
	Analytics: &AnalyticsConfig{
		MaxEvents: 100 << 10,
		MaxAge:    10 * 60, // seconds
	},
	TwoFactorExpiration: 10 * 60, // seconds
}

func (c *Config) Validate() {
	var errors []string
	if c.ContentBucket == "" {
		errors = append(errors, "ContentBucket not set")
	}
	if len(c.Storage) == 0 {
		errors = append(errors, "No storage configs set")
	}
	if c.Stripe == nil || c.Stripe.SecretKey == "" || c.Stripe.PublishableKey == "" {
		errors = append(errors, "No stripe key set")
	}
	if len(c.SecretSignatureKeys) == 0 {
		errors = append(errors, "No secret signature keys")
	}

	if !c.Debug {
		if c.TLSCert == "" {
			errors = append(errors, "TLSCert not set")
		}
		if c.TLSKey == "" {
			errors = append(errors, "TLSKey not set")
		}
	}
	if c.StaticResourceURL == "" {
		if os.Getenv("GOPATH") == "" {
			errors = append(errors, "StaticResourceURL not set")
		} else {
			// In dev we can use a local file server in the app
			c.StaticResourceURL = fmt.Sprintf("https://%s/static", c.WebDomain)
		}
	} else if n := len(c.StaticResourceURL); c.StaticResourceURL[n-1] == '/' {
		c.StaticResourceURL = c.StaticResourceURL[:n-1]
	}
	c.StaticResourceURL = strings.Replace(c.StaticResourceURL, "{BuildNumber}", config.BuildNumber, -1)
	if len(errors) != 0 {
		fmt.Fprintf(os.Stderr, "Config failed validation:\n")
		for _, e := range errors {
			fmt.Fprintf(os.Stderr, "- %s\n", e)
		}
		os.Exit(1)
	}
}
