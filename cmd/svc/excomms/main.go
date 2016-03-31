package main

import (
	"flag"

	"github.com/sprucehealth/backend/boot"
	"github.com/sprucehealth/backend/libs/conc"
)

var config struct {
	excommsServicePort      int
	excommsAPIURL           string
	directoryServiceURL     string
	settingsServiceURL      string
	twilioAuthToken         string
	twilioAccountSID        string
	twilioApplicationSID    string
	sendgridAPIKey          string
	attachmentBucket        string
	attachmentPrefix        string
	externalMessageTopic    string
	eventTopic              string
	incomingRawMessageQueue string
	dbHost                  string
	dbPassword              string
	dbName                  string
	dbUserName              string
	dbPort                  int
	dbCACert                string
	dbTLS                   string
	httpAddr                string
	proxyProtocol           bool
	excommsServiceURL       string
	incomingRawMessageTopic string
	kmsKeyARN               string
	resourceCleanerQueueURL string
	resourceCleanerTopic    string
	segmentIOKey            string
}

func init() {
	flag.IntVar(&config.excommsServicePort, "excomms_port", 5200, "port on which excomms service should listen")
	flag.StringVar(&config.excommsAPIURL, "excommsapi_endpoint", "", "url for excomms api")
	flag.StringVar(&config.twilioAccountSID, "twilio_account_sid", "", "account sid for twilio account")
	flag.StringVar(&config.twilioApplicationSID, "twilio_application_sid", "", "application sid for twilio")
	flag.StringVar(&config.twilioAuthToken, "twilio_auth_token", "", "auth token for twilio account")
	flag.StringVar(&config.directoryServiceURL, "directory_endpoint", "", "url to connect with directory service")
	flag.StringVar(&config.settingsServiceURL, "settings_endpoint", "", "url to connect with settings service")
	flag.StringVar(&config.sendgridAPIKey, "sendgrid_api_key", "", "sendgrid api key")
	flag.StringVar(&config.externalMessageTopic, "sns_external_message_topic", "", "sns topic on which to post external message events")
	flag.StringVar(&config.eventTopic, "sns_event_topic", "", "SNS topic on which to publish events")
	flag.StringVar(&config.dbHost, "db_host", "", "database host")
	flag.StringVar(&config.dbPassword, "db_password", "", "database password")
	flag.StringVar(&config.dbName, "db_name", "", "database name")
	flag.StringVar(&config.dbUserName, "db_username", "", "database username")
	flag.IntVar(&config.dbPort, "db_port", 3306, "database port")
	flag.StringVar(&config.dbCACert, "db_ca_cert", "", "Path to database CA certificate")
	flag.StringVar(&config.dbTLS, "db_tls", "skip-verify", "Enable TLS for database connection (one of 'true', 'false', 'skip-verify'). Ignored if CA cert provided.")
	flag.StringVar(&config.incomingRawMessageQueue, "queue_incoming_raw_message", "", "queue name for receiving incoming raw messages")
	flag.StringVar(&config.httpAddr, "http", "0.0.0.0:8900", "listen for http on `host:port`")
	flag.BoolVar(&config.proxyProtocol, "proxyproto", false, "enabled proxy protocol")
	flag.StringVar(&config.excommsServiceURL, "excomms_url", "localhost:5200", "url for events processor service. format `host:port`")
	flag.StringVar(&config.incomingRawMessageTopic, "sns_incoming_raw_message_topic", "", "Inbound msg topic")
	flag.StringVar(&config.attachmentBucket, "s3_attachment_bucket", "dev-baymax-storage", "bucket name for s3 storage")
	flag.StringVar(&config.attachmentPrefix, "s3_attachment_prefix", "excomms-media", "prefix for excomms media attachments")
	flag.StringVar(&config.kmsKeyARN, "kms_key_arn", "", "the arn of the master key that should be used to encrypt outbound and decrypt inbound data")
	flag.StringVar(&config.resourceCleanerTopic, "sns_resource_cleaner_topic", "", "sns topic for publishing requests to delete resources")
	flag.StringVar(&config.resourceCleanerQueueURL, "resource_cleaner_queue_url", "", "sqs queue that contains requests to delete resources")
	flag.StringVar(&config.segmentIOKey, "segmentio_key", "", "Segment IO API `key`")
}

func main() {
	bootSvc := boot.NewService("excomms")

	conc.Go(func() {
		runAPI(bootSvc)
	})

	conc.Go(func() {
		runService(bootSvc)
	})

	boot.WaitForTermination()
}
