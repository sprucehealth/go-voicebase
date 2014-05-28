package main

import (
	"carefront/common/config"
	"fmt"
	"os"
)

type TwilioConfig struct {
	AccountSid string `long:"twilio_account_sid" description:"Twilio AccountSid"`
	AuthToken  string `long:"twilio_auth_token" description:"Twilio AuthToken"`
	FromNumber string `long:"twilio_from_number" description:"Twilio From Number for Messages"`
}

type DosespotConfig struct {
	ClinicId  int64  `long:"clinic_id" description:"Clinic Id for dosespot"`
	ClinicKey string `long:"clinic_key" description:"Clinic Key for dosespot"`
	UserId    int64  `long:"user_id" description:"User Id for dosespot"`
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

type Config struct {
	*config.BaseConfig
	ProxyProtocol            bool                        `long:"proxy_protocol" description:"Enable if behind a proxy that uses the PROXY protocol"`
	ListenAddr               string                      `short:"l" long:"listen" description:"Address and port on which to listen (e.g. 127.0.0.1:8080)"`
	TLSListenAddr            string                      `long:"tls_listen" description:"Address and port on which to listen (e.g. 127.0.0.1:8080)"`
	TLSCert                  string                      `long:"tls_cert" description:"Path of SSL certificate"`
	TLSKey                   string                      `long:"tls_key" description:"Path of SSL private key"`
	InfoAddr                 string                      `long:"info_addr" description:"Address to listen on for the info server"`
	DB                       *config.DB                  `group:"Database" toml:"database"`
	MaxInMemoryForPhotoMB    int64                       `long:"max_in_memory_photo" description:"Amount of data in MB to be held in memory when parsing multipart form data"`
	ContentBucket            string                      `long:"content_bucket" description:"S3 Bucket name for all static content"`
	CaseBucket               string                      `long:"case_bucket" description:"S3 Bucket name for case information"`
	PatientLayoutBucket      string                      `long:"client_layout_bucket" description:"S3 Bucket name for client digestable layout for patient information intake"`
	VisualLayoutBucket       string                      `long:"patient_layout_bucket" description:"S3 Bucket name for human readable layout for patient information intake"`
	DoctorVisualLayoutBucket string                      `long:"doctor_visual_layout_bucket" description:"S3 Bucket name for patient overview for doctor's viewing"`
	DoctorLayoutBucket       string                      `long:"doctor_layout_bucket" description:"S3 Bucket name for pre-processed patient overview for doctor's viewing"`
	PhotoBucket              string                      `long:"photo_bucket" description:"S3 Bucket name for uploaded photos"`
	Debug                    bool                        `long:"debug" description:"Enable debugging"`
	DoseSpotUserId           string                      `long:"dose_spot_user_id" description:"DoseSpot UserId for eRx integration"`
	NoServices               bool                        `long:"noservices" description:"Disable connecting to remote services"`
	ERxRouting               bool                        `long:"erx_routing" description:"Disable sending of prescriptions electronically"`
	ERxQueue                 string                      `long:"erx_queue" description:"Erx queue name"`
	AuthTokenExpiration      int                         `long:"auth_token_expire" description:"Expiration time in seconds for the auth token"`
	AuthTokenRenew           int                         `long:"auth_token_renew" description:"Time left below which to renew the auth token"`
	StaticContentBaseUrl     string                      `long:"static_content_base_url" description:"URL from which to serve static content"`
	Twilio                   *TwilioConfig               `group:"Twilio" toml:"twilio"`
	DoseSpot                 *DosespotConfig             `group:"Dosespot" toml:"dosespot"`
	SmartyStreets            *SmartyStreetsConfig        `group:"smarty_streets" toml:"smarty_streets"`
	StripeSecretKey          string                      `long:"strip_secret_key" description:"Stripe secret key"`
	IOSDeeplinkScheme        string                      `long:"ios_deeplink_scheme" description:"Scheme for iOS deep-links (e.g. spruce://)"`
	NotifiyConfigs           *config.NotificationConfigs `group:"notification" toml:"notification"`
	Analytics                *AnalyticsConfig            `group:"Analytics" toml:"analytics"`
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
	ListenAddr:            ":8080",
	TLSListenAddr:         ":8443",
	InfoAddr:              ":9000",
	CaseBucket:            "carefront-cases",
	MaxInMemoryForPhotoMB: defaultMaxInMemoryPhotoMB,
	AuthTokenExpiration:   60 * 60 * 24 * 2,
	AuthTokenRenew:        60 * 60 * 36,
	IOSDeeplinkScheme:     "spruce",
	Analytics: &AnalyticsConfig{
		MaxEvents: 100 << 10,
		MaxAge:    10 * 60, // seconds
	},
}

func (c *Config) Validate() {
	var errors []string
	if c.ContentBucket == "" {
		errors = append(errors, "ContentBucket not set")
	}
	if c.VisualLayoutBucket == "" {
		errors = append(errors, "VisualLayoutBucket not set")
	}
	if c.PatientLayoutBucket == "" {
		errors = append(errors, "PatientLayoutBucket not set")
	}
	if c.DoctorVisualLayoutBucket == "" {
		errors = append(errors, "DoctorVisualLayoutBucket not set")
	}
	if c.DoctorLayoutBucket == "" {
		errors = append(errors, "DoctorLayoutBucket")
	}
	if c.PhotoBucket == "" {
		errors = append(errors, "PhotoBucket not set")
	}
	if len(errors) != 0 {
		fmt.Fprintf(os.Stderr, "Config failed validation:\n")
		for _, e := range errors {
			fmt.Fprintf(os.Stderr, "- %s\n", e)
		}
		os.Exit(1)
	}
}
