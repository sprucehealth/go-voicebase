package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sprucehealth/backend/boot"
	"github.com/sprucehealth/backend/libs/conc"
	"github.com/sprucehealth/backend/libs/golog"
)

var config struct {
	excommsServicePort      int
	excommsAPIURL           string
	directoryServiceURL     string
	twilioAuthToken         string
	twilioAccountSID        string
	twilioApplicationSID    string
	sendgridAPIKey          string
	awsRegion               string
	awsAccessKey            string
	awsSecretKey            string
	attachmentBucket        string
	attachmentPrefix        string
	externalMessageTopic    string
	incomingRawMessageQueue string
	debug                   bool
	dbHost                  string
	dbPassword              string
	dbName                  string
	dbUserName              string
	dbPort                  int
	httpAddr                string
	proxyProtocol           bool
	excommsServiceURL       string
	incomingRawMessageTopic string
	env                     string
}

func init() {
	flag.IntVar(&config.excommsServicePort, "excomms_port", 5200, "port on which excomms service should listen")
	flag.StringVar(&config.excommsAPIURL, "excommsapi_endpoint", "", "url for excomms api")
	flag.StringVar(&config.twilioAccountSID, "twilio_account_sid", "", "account sid for twilio account")
	flag.StringVar(&config.twilioApplicationSID, "twilio_application_sid", "", "application sid for twilio")
	flag.StringVar(&config.twilioAuthToken, "twilio_auth_token", "", "auth token for twilio account")
	flag.StringVar(&config.directoryServiceURL, "directory_endpoint", "", "url to connect with directory service")
	flag.StringVar(&config.awsRegion, "aws_region", "us-east-1", "aws region")
	flag.StringVar(&config.awsAccessKey, "aws_access_key", "", "access key for aws user")
	flag.StringVar(&config.awsSecretKey, "aws_secret_key", "", "secret key for aws user")
	flag.StringVar(&config.sendgridAPIKey, "sendgrid_api_key", "", "sendgrid api key")
	flag.StringVar(&config.externalMessageTopic, "sns_external_message_topic", "", "sns topic on which to post external message events")
	flag.BoolVar(&config.debug, "debug", false, "debug flag")
	flag.StringVar(&config.dbHost, "db_host", "", "database host")
	flag.StringVar(&config.dbPassword, "db_password", "", "database password")
	flag.StringVar(&config.dbName, "db_name", "", "database name")
	flag.StringVar(&config.dbUserName, "db_username", "", "database username")
	flag.IntVar(&config.dbPort, "db_port", 3306, "database port")
	flag.StringVar(&config.incomingRawMessageQueue, "queue_incoming_raw_message", "", "queue name for receiving incoming raw messages")
	flag.StringVar(&config.httpAddr, "http", "0.0.0.0:8900", "listen for http on `host:port`")
	flag.BoolVar(&config.proxyProtocol, "proxyproto", false, "enabled proxy protocol")
	flag.StringVar(&config.excommsServiceURL, "excomms_url", "localhost:5200", "url for events processor service. format `host:port`")
	flag.StringVar(&config.incomingRawMessageTopic, "sns_incoming_raw_message_topic", "", "Inbound msg topic")
	flag.StringVar(&config.env, "env", "dev", "environment")
	flag.StringVar(&config.attachmentBucket, "s3_attachment_bucket", "dev-baymax-storage", "bucket name for s3 storage")
	flag.StringVar(&config.attachmentPrefix, "s3_attachment_prefix", "excomms-media", "prefix for excomms media attachments")
}

func main() {
	boot.ParseFlags("EXCOMMS_")

	conc.Go(func() {
		runAPI()
	})

	conc.Go(func() {
		runService()
	})

	// Wait for an external process interrupt to quit the program
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, os.Kill, syscall.SIGTERM)
	select {
	case sig := <-sigCh:
		golog.Infof("Quitting due to signal %s", sig.String())
		break
	}

}