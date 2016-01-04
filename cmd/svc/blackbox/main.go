package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/sprucehealth/backend/boot"
	auth_blackbox "github.com/sprucehealth/backend/cmd/svc/auth/blackbox"
	"github.com/sprucehealth/backend/cmd/svc/blackbox/harness"
	"github.com/sprucehealth/backend/cmd/svc/blackbox/internal/dal"
	directory_blackbox "github.com/sprucehealth/backend/cmd/svc/directory/blackbox"
	"github.com/sprucehealth/backend/libs/dbutil"
	"github.com/sprucehealth/backend/libs/errors"
	"github.com/sprucehealth/backend/libs/golog"
)

var config struct {
	listenPort          int64
	debug               bool
	noDB                bool
	dbHost              string
	dbPort              int64
	dbName              string
	dbUser              string
	dbPassword          string
	dbCACert            string
	dbTLSCert           string
	dbTLSKey            string
	suites              string
	tests               string
	configJSON          string
	suiteStagger        int
	suiteRepeat         int
	testStagger         int
	maxTestParallel     int
	runOnce             bool
	slackWebhook        string
	slackWebhookChannel string
}

func init() {
	flag.Int64Var(&config.listenPort, "rpc.listen.port", 50051, "the port on which to listen for rpc call")
	flag.BoolVar(&config.debug, "debug", false, "enables golog debug logging for the application")
	flag.BoolVar(&config.noDB, "no.db", false, "no db diables the DAL functionality related to run result record storing")
	flag.StringVar(&config.dbHost, "db.host", "localhost", "the host at which we should attempt to connect to the database")
	flag.Int64Var(&config.dbPort, "db.port", 3306, "the port on which we should attempt to connect to the database")
	flag.StringVar(&config.dbName, "db.name", "blackbox", "the name of the database which we should connect to")
	flag.StringVar(&config.dbUser, "db.user", "blackbox", "the name of the user we should connext to the database as")
	flag.StringVar(&config.dbPassword, "db.password", "blackbox", "the password we should use when connecting to the database")
	flag.StringVar(&config.dbCACert, "db.ca.cert", "", "the ca cert to use when connecting to the database")
	flag.StringVar(&config.dbTLSCert, "db.tls.cert", "", "the tls cert to use when connecting to the database")
	flag.StringVar(&config.dbTLSKey, "db.tls.key", "", "the tls key to use when connecting to the database")
	flag.StringVar(&config.suites, "suites", "", "A comma seperated list of which test suites to run")
	flag.StringVar(&config.tests, "tests", "", "A comma seperated list of which test to run")
	flag.StringVar(&config.configJSON, "config", "", "A JSON representation of the global config to make available to tests")
	flag.IntVar(&config.suiteStagger, "suite.stagger", 2, "A measure in seconds of how long to delay between beginning each parallel suite execution")
	flag.IntVar(&config.suiteRepeat, "suite.repeat", 60, "A measure in seconds of how long to delay between iterative runs of the test system")
	flag.IntVar(&config.testStagger, "test.stagger", 1, "A measure in seconds of how long to delay between beginning each parallel test execution")
	flag.IntVar(&config.maxTestParallel, "max.test.parallel", 10, "The maximum number of tests to execute in parallel")
	flag.BoolVar(&config.runOnce, "run.once", false, "Set this flag to only run each suite/test once and then exit")
	flag.StringVar(&config.slackWebhook, "slack.webhook", "", "The webhook to use as a report processor")
	flag.StringVar(&config.slackWebhookChannel, "slack.webhook.channel", "x-blackbox", "The channel the webhook should report to")
}

func main() {
	boot.ParseFlags("BLACK_BOX_SERVICE_")
	configureLogging()
	loadTestConfig()
	configureWebhook()

	golog.Infof("Initializing database connection on %s:%d, user: %s, db: %s...", config.dbHost, config.dbPort, config.dbUser, config.dbName)
	conn, err := dbutil.ConnectMySQL(&dbutil.DBConfig{
		Host:     config.dbHost,
		Port:     config.dbPort,
		Name:     config.dbName,
		User:     config.dbUser,
		Password: config.dbPassword,
		CACert:   config.dbCACert,
		TLSCert:  config.dbTLSCert,
		TLSKey:   config.dbTLSKey,
	})
	if err != nil {
		golog.Fatalf("failed to iniitlize db connection: %s", err)
	}

	if !config.noDB {
		harness.SetDAL(dal.New(conn))
	}

	// Register the different test suites
	registrationConfig := &harness.RegistrationConfig{SuitesToRegister: parseSuiteNames(), TestsToRegister: parseTestNames()}
	harness.Register(auth_blackbox.NewTests(), registrationConfig)
	harness.Register(directory_blackbox.NewTests(), registrationConfig)
	harness.Execute(&harness.ExecutionConfig{
		SuiteStagger:    time.Duration(config.suiteStagger) * time.Second,
		SuiteRepeat:     time.Duration(config.suiteRepeat) * time.Second,
		TestStagger:     time.Duration(config.testStagger) * time.Second,
		MaxTestParallel: config.maxTestParallel,
		RunOnce:         config.runOnce,
	})
}

func loadTestConfig() {
	if config.configJSON == "" {
		return
	}
	var testConfig map[string]string
	if err := json.Unmarshal([]byte(config.configJSON), &testConfig); err != nil {
		golog.Fatalf("Failed to parse config JSON: %s", err.Error())
	}
	for k, v := range testConfig {
		harness.SetConfig(k, v)
	}
}

func parseSuiteNames() map[string]struct{} {
	suites := strings.Split(config.suites, ",")
	parsedSuites := make(map[string]struct{}, len(suites))
	for _, s := range suites {
		if s != "" {
			parsedSuites[s] = struct{}{}
		}
	}
	return parsedSuites
}

func parseTestNames() map[string]struct{} {
	suites := strings.Split(config.tests, ",")
	parsedTests := make(map[string]struct{}, len(suites))
	for _, s := range suites {
		if s != "" {
			parsedTests[s] = struct{}{}
		}
	}
	return parsedTests
}

func configureLogging() {
	if config.debug {
		golog.Default().SetLevel(golog.DEBUG)
		golog.Debugf("Debug logging enabled...")
	}
}

type slackWebhookInput struct {
	Text      string `json:"text"`
	Username  string `json:"username"`
	IconEmoji string `json:"icon_emoji"`
	IconURL   string `json:"icon_url"`
	Channel   string `json:"channel"`
}

func configureWebhook() {
	if config.slackWebhook != "" {
		harness.RegisterReportProcessor(func(r *harness.SuiteRunReport) error {
			input := &slackWebhookInput{
				Text:      "```\n" + r.String() + "\n```",
				Username:  "blackbox",
				IconEmoji: ":batman:",
				Channel:   config.slackWebhookChannel,
			}
			data, err := json.Marshal(input)
			if err != nil {
				return errors.Trace(err)
			}
			resp, err := http.DefaultClient.Post(config.slackWebhook, "application/json", bytes.NewReader(data))
			defer resp.Body.Close()
			if err != nil {
				return errors.Trace(err)
			}
			if resp.StatusCode != http.StatusOK {
				d, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return errors.Trace(err)
				}
				return errors.Trace(fmt.Errorf("%d: %s", resp.StatusCode, string(d)))
			}
			return nil
		})
	}
}
