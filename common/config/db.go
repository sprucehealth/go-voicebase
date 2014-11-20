package config

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sprucehealth/backend/Godeps/_workspace/src/github.com/go-sql-driver/mysql"
	_ "github.com/sprucehealth/backend/Godeps/_workspace/src/github.com/lib/pq"
)

type DB struct {
	User     string `description:"Username for accessing database"`
	Password string `description:"Password for accessing database"`
	Host     string `description:"Database host"`
	Port     int    `description:"Database port"`
	Name     string `description:"Database name"`
	CACert   string `description:"Database TLS CA certificate path"`
	TLSCert  string `description:"Database TLS client certificate path"`
	TLSKey   string `description:"Database TLS client key path"`
}

func (c *DB) ConnectMySQL(bconf *BaseConfig) (*sql.DB, error) {
	if c.User == "" || c.Host == "" || c.Name == "" {
		return nil, errors.New("missing one or more of user, host, or name for db config")
	}

	enableTLS := c.CACert != "" && c.TLSCert != "" && c.TLSKey != ""
	if enableTLS {
		rootCertPool := x509.NewCertPool()
		pem, err := bconf.ReadURI(c.CACert)
		if err != nil {
			return nil, err
		}
		if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
			return nil, fmt.Errorf("Failed to append PEM.")
		}
		clientCert := make([]tls.Certificate, 0, 1)
		cert, err := bconf.ReadURI(c.TLSCert)
		if err != nil {
			return nil, err
		}
		key, err := bconf.ReadURI(c.TLSKey)
		if err != nil {
			return nil, err
		}
		certs, err := tls.X509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}
		clientCert = append(clientCert, certs)
		mysql.RegisterTLSConfig("custom", &tls.Config{
			RootCAs:            rootCertPool,
			Certificates:       clientCert,
			InsecureSkipVerify: true,
		})
	}

	tlsOpt := "?parseTime=true"
	if enableTLS {
		tlsOpt += "&tls=custom"
	}
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s%s&charset=utf8mb4&collation=utf8mb4_unicode_ci", c.User, c.Password, c.Host, c.Port, c.Name, tlsOpt))
	if err != nil {
		return nil, err
	}
	// test the connection to the database by running a ping against it
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func (c *DB) ConnectPostgres() (*sql.DB, error) {
	dbArgs := fmt.Sprintf("host=%s port=%d dbname=%s sslmode=%s", c.Host, c.Port, c.Name, "require")
	if c.User != "" {
		dbArgs += " user=" + c.User
	}
	if c.Password != "" {
		dbArgs += " password=" + c.Password
	}
	db, err := sql.Open("postgres", dbArgs)
	if err != nil {
		return nil, err
	}
	// Make sure the database connection is working
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
