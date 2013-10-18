package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"carefront/api"
	_ "github.com/go-sql-driver/mysql"
)

var (
	flagListenAddr   = flag.String("listen", ":8080", "Address and port to listen on")
	flagCertLocation = flag.String("cert_key", "cert.pem", "Location of certificate for SSL")
	flagKeyLocation  = flag.String("private_key", "key.pem", "Location of key for SSL")
	flagS3CaseBucket = flag.String("case_bucket", "carefront-cases", "Bucket name holding case information on S3")
	flagAWSSecretKey = flag.String("aws_secret_key", "", "AWS Secret Key for uploading files to S3")
	flagAWSAccessKey = flag.String("aws_access_key", "", "AWS Access Key to upload files to S3")
	flagDBUser       = flag.String("db_user", "", "Username for accessing database")
	flagDBPassword   = flag.String("db_password", "", "Password for accessing database")
	flagDBHost       = flag.String("db_host", "", "Database host url")
	flagDBName       = flag.String("db_name", "", "Database name on database server")
)

const (
	CertKeyLocation    string = "CERT_KEY"
	PrivateKeyLocation string = "PRIVATE_KEY"
)

func parseFlags() {
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "@") {
		f, err := os.Open(os.Args[1][1:])
		if err == nil {
			argBytes, err := ioutil.ReadAll(f)
			f.Close()
			if err == nil {
				args := strings.Split(strings.TrimSpace(string(argBytes)), "\n")
				filteredArgs := make([]string, 0, len(args))
				for _, a := range args {
					if !strings.HasPrefix(a, "#") {
						filteredArgs = append(filteredArgs, a)
					}
				}
				os.Args = append(append(os.Args[:1], filteredArgs...), os.Args[2:]...)
			}
		}
	}
	flag.VisitAll(func(fl *flag.Flag) {
		val := os.Getenv("arg_" + strings.Replace(fl.Name, ".", "_", -1))
		if val != "" {
			fl.Value.Set(val)
		}
	})
	flag.Parse()
}

func main() {
	parseFlags()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", *flagDBUser, *flagDBPassword, *flagDBHost, *flagDBName)

	// this gives us a connection pool to the sql instance
	// without executing any statements against the sql database
	// or checking the network connection and authentication to the database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// test the connection to the database by running a ping against it
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	authApi := &api.AuthService{db}
	dataApi := &api.DataService{db}
	mux := &AuthServeMux{*http.NewServeMux(), authApi}

	authHandler := &AuthenticationHandler{authApi}
	pingHandler := PingHandler(0)
	photoHandler := &PhotoUploadHandler{&api.PhotoService{*flagAWSAccessKey, *flagAWSSecretKey}, *flagS3CaseBucket, dataApi}
	getSignedUrlsHandler := &GetSignedUrlsHandler{&api.PhotoService{*flagAWSAccessKey, *flagAWSSecretKey}}

	mux.Handle("/v1/authenticate", authHandler)
	mux.Handle("/v1/signup", authHandler)
	mux.Handle("/v1/logout", authHandler)
	mux.Handle("/v1/ping", pingHandler)
	mux.Handle("/v1/upload", photoHandler)
	mux.Handle("/v1/imagesForCase/", getSignedUrlsHandler)

	s := &http.Server{
		Addr:           *flagListenAddr,
		Handler:        mux,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServeTLS(*flagCertLocation, *flagKeyLocation))
}
