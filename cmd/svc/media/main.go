package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/grpc"

	"github.com/sprucehealth/backend/boot"
	"github.com/sprucehealth/backend/cmd/svc/media/internal/dal"
	"github.com/sprucehealth/backend/cmd/svc/media/internal/handlers"
	"github.com/sprucehealth/backend/libs/clock"
	"github.com/sprucehealth/backend/libs/dbutil"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/libs/mux"
	"github.com/sprucehealth/backend/libs/sig"
	"github.com/sprucehealth/backend/libs/urlutil"
	"github.com/sprucehealth/backend/shttputil"
	"github.com/sprucehealth/backend/svc/auth"
)

var (
	flagHTTPListenAddr     = flag.String("http_listen_addr", ":8081", "host:port to listen on for http requests")
	flagWebDomain          = flag.String("web_domain", "", "Web `domain`")
	flagMediaAPIDomain     = flag.String("media_api_domain", "", "Media API `domain`")
	flagMediaStorageBucket = flag.String("media_storage_bucket", "", "storage bucket for media")
	flagSigKeys            = flag.String("signature_keys_csv", "", "csv signature keys")
	flagBehindProxy        = flag.Bool("behind_proxy", false, "Flag to indicate when the service is behind a proxy")

	// Services
	flagAuthAddr = flag.String("auth_addr", "", "host:port of auth service")

	// DB
	flagDBHost     = flag.String("db_host", "localhost", "the host at which we should attempt to connect to the database")
	flagDBPort     = flag.Int("db_port", 3306, "the port on which we should attempt to connect to the database")
	flagDBName     = flag.String("db_name", "media", "the name of the database which we should connect to")
	flagDBUser     = flag.String("db_user", "baymax-media", "the name of the user we should connext to the database as")
	flagDBPassword = flag.String("db_password", "baymax-media", "the password we should use when connecting to the database")
	flagDBCACert   = flag.String("db_ca_cert", "", "the ca cert to use when connecting to the database")
	flagDBTLSCert  = flag.String("db_tls_cert", "", "the tls cert to use when connecting to the database")
	flagDBTLSKey   = flag.String("db_tls_key", "", "the tls key to use when connecting to the database")
)

func main() {
	svc := boot.NewService("media")
	awsSession, err := svc.AWSSession()
	if err != nil {
		golog.Fatalf("Failed to create AWS session: %s", err)
	}

	if *flagMediaAPIDomain == "" {
		golog.Fatalf("Media API Domain not specified")
	}

	if *flagMediaStorageBucket == "" {
		golog.Fatalf("Media Storage bucket not specified")
	}

	if *flagAuthAddr == "" {
		golog.Fatalf("Auth service addr not configured")
	}
	conn, err := grpc.Dial(*flagAuthAddr, grpc.WithInsecure())
	if err != nil {
		golog.Fatalf("Unable to connect to auth service: %s", err)
	}
	authClient := auth.NewAuthClient(conn)

	golog.Infof("Initializing database connection on %s:%d, user: %s, db: %s...", *flagDBHost, *flagDBPort, *flagDBUser, *flagDBName)
	db, err := dbutil.ConnectMySQL(&dbutil.DBConfig{
		Host:     *flagDBHost,
		Port:     *flagDBPort,
		Name:     *flagDBName,
		User:     *flagDBUser,
		Password: *flagDBPassword,
		CACert:   *flagDBCACert,
		TLSCert:  *flagDBTLSCert,
		TLSKey:   *flagDBTLSKey,
	})
	if err != nil {
		golog.Fatalf("Failed to initialize DB connection: %s", err)
	}

	if *flagSigKeys == "" {
		golog.Fatalf("signature_keys_csv required")
	}
	sigKeys := strings.Split(*flagSigKeys, ",")
	sigKeysByteSlice := make([][]byte, len(sigKeys))
	for i, sk := range sigKeys {
		sigKeysByteSlice[i] = []byte(sk)
	}
	signer, err := sig.NewSigner(sigKeysByteSlice, nil)
	if err != nil {
		golog.Fatalf("Failed to create signer: %s", err.Error())
	}

	r := mux.NewRouter()
	handlers.InitRoutes(r,
		awsSession,
		authClient,
		urlutil.NewSigner("https://"+*flagMediaAPIDomain, signer, clock.New()),
		dal.New(db),
		*flagWebDomain)
	h := httputil.LoggingHandler(r, "media", *flagBehindProxy, nil)

	fmt.Printf("HTTP Listening on %s\n", *flagHTTPListenAddr)
	server := &http.Server{
		Addr:           *flagHTTPListenAddr,
		Handler:        httputil.FromContextHandler(shttputil.CompressResponse(h, httputil.CompressResponse)),
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		server.ListenAndServe()
	}()

	boot.WaitForTermination()
}