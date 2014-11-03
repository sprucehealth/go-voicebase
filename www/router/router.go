package router

import (
	"database/sql"
	"io"
	"net/http"

	"github.com/sprucehealth/backend/analytics"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/email"
	"github.com/sprucehealth/backend/environment"
	"github.com/sprucehealth/backend/libs/dispatch"
	"github.com/sprucehealth/backend/libs/erx"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/httputil"
	"github.com/sprucehealth/backend/libs/storage"
	"github.com/sprucehealth/backend/libs/stripe"
	"github.com/sprucehealth/backend/medrecord"
	"github.com/sprucehealth/backend/passreset"
	resources "github.com/sprucehealth/backend/third_party/github.com/cookieo9/resources-go"
	"github.com/sprucehealth/backend/third_party/github.com/gorilla/context"
	"github.com/sprucehealth/backend/third_party/github.com/gorilla/mux"
	"github.com/sprucehealth/backend/third_party/github.com/samuel/go-metrics/metrics"
	"github.com/sprucehealth/backend/www"
	"github.com/sprucehealth/backend/www/admin"
	"github.com/sprucehealth/backend/www/dronboard"
	"github.com/sprucehealth/backend/www/home"
)

var robotsTXT = []byte(`Sitemap: https://www.sprucehealth.com/sitemap.xml
User-agent: *
Disallow: /login
`)

var sitemapXML = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	<url>
		<loc>https://www.sprucehealth.com</loc>
		<changefreq>daily</changefreq>
	</url>
	<url>
		<loc>https://www.sprucehealth.com/meet-the-doctors</loc>
		<changefreq>daily</changefreq>
	</url>
	<url>
		<loc>https://www.sprucehealth.com/about</loc>
		<changefreq>daily</changefreq>
	</url>
	<url>
		<loc>https://www.sprucehealth.com/contact</loc>
		<changefreq>daily</changefreq>
	</url>
</urlset>
`)

type Config struct {
	DataAPI              api.DataAPI
	AuthAPI              api.AuthAPI
	SMSAPI               api.SMSAPI
	ERxAPI               erx.ERxAPI
	Dispatcher           *dispatch.Dispatcher
	AnalyticsDB          *sql.DB
	AnalyticsLogger      analytics.Logger
	FromNumber           string
	EmailService         email.Service
	SupportEmail         string
	WebDomain            string
	StaticResourceURL    string
	StripeCli            *stripe.StripeService
	Signer               *common.Signer
	Stores               map[string]storage.Store
	WebPassword          string
	TemplateLoader       *www.TemplateLoader
	MetricsRegistry      metrics.Registry
	OnboardingURLExpires int64
	TwoFactorExpiration  int
}

func New(c *Config) http.Handler {
	if c.StaticResourceURL == "" {
		c.StaticResourceURL = "/static"
	}

	c.TemplateLoader.MustLoadTemplate("base.html", "", map[string]interface{}{
		"staticURL": func(path string) string {
			return c.StaticResourceURL + path
		},
		"isEnv": func(env string) bool {
			return environment.GetCurrent() == env
		},
	})

	router := mux.NewRouter().StrictSlash(true)
	router.KeepContext = true
	router.Handle("/login", www.NewLoginHandler(c.AuthAPI, c.SMSAPI, c.FromNumber, c.TwoFactorExpiration, c.TemplateLoader, c.MetricsRegistry.Scope("login")))
	router.Handle("/login/verify", www.NewLoginVerifyHandler(c.AuthAPI, c.TemplateLoader, c.MetricsRegistry.Scope("login-verify")))
	router.Handle("/logout", www.NewLogoutHandler(c.AuthAPI))
	router.Handle("/robots.txt", RobotsTXTHandler())
	router.Handle("/sitemap.xml", SitemapXMLHandler())
	router.PathPrefix("/static").Handler(http.StripPrefix("/static", http.FileServer(www.ResourceFileSystem)))

	router.Handle("/privacy", StaticHTMLHandler("terms.html"))
	router.Handle("/medication-affordability", StaticHTMLHandler("medafford.html"))

	home.SetupRoutes(router, c.DataAPI, c.AuthAPI, c.WebPassword, c.AnalyticsLogger, c.TemplateLoader, c.MetricsRegistry.Scope("home"))
	passreset.SetupRoutes(router, c.DataAPI, c.AuthAPI, c.SMSAPI, c.FromNumber, c.EmailService, c.SupportEmail, c.WebDomain, c.TemplateLoader, c.MetricsRegistry.Scope("reset-password"))
	dronboard.SetupRoutes(router, &dronboard.Config{
		DataAPI:         c.DataAPI,
		AuthAPI:         c.AuthAPI,
		SMSAPI:          c.SMSAPI,
		SMSFromNumber:   c.FromNumber,
		SupportEmail:    c.SupportEmail,
		Dispatcher:      c.Dispatcher,
		StripeCli:       c.StripeCli,
		Signer:          c.Signer,
		Stores:          c.Stores,
		TemplateLoader:  c.TemplateLoader,
		MetricsRegistry: c.MetricsRegistry.Scope("doctor-onboard"),
	})
	admin.SetupRoutes(router, &admin.Config{
		DataAPI:              c.DataAPI,
		AuthAPI:              c.AuthAPI,
		ERxAPI:               c.ERxAPI,
		AnalyticsDB:          c.AnalyticsDB,
		Signer:               c.Signer,
		Stores:               c.Stores,
		TemplateLoader:       c.TemplateLoader,
		EmailService:         c.EmailService,
		OnboardingURLExpires: c.OnboardingURLExpires,
		MetricsRegistry:      c.MetricsRegistry.Scope("admin"),
	})

	patientAuthFilter := www.AuthRequiredFilter(c.AuthAPI, []string{api.PATIENT_ROLE}, nil)
	router.Handle("/patient/medical-record", patientAuthFilter(medrecord.NewWebDownloadHandler(c.DataAPI, c.Stores["medicalrecords"])))
	router.Handle("/patient/medical-record/media/{media:[0-9]+}", patientAuthFilter(medrecord.NewPhotoHandler(c.DataAPI, c.Stores["media"], c.Signer)))

	secureRedirectHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Forwarded-Proto") != "https" {
			u := r.URL
			u.Scheme = "https"
			u.Host = r.Host
			http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
			return
		}
		router.ServeHTTP(w, r)
	})
	return httputil.MetricsHandler(
		httputil.CompressResponse(
			httputil.DecompressRequest(
				context.ClearHandler(
					httputil.RequestIDHandler(
						httputil.LoggingHandler(
							secureRedirectHandler,
							golog.Default()))))),
		c.MetricsRegistry)
}

func StaticHTMLHandler(name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := resources.Open("templates/static/" + name)
		if err != nil {
			www.InternalServerError(w, r, err)
		}
		defer f.Close()
		// TODO: set cache headers
		r.Header.Set("Content-Type", "text/html")
		io.Copy(w, f)
	})
}

func RobotsTXTHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		// TODO: set cache headers
		if _, err := w.Write(robotsTXT); err != nil {
			golog.Errorf(err.Error())
		}
	})
}

func SitemapXMLHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		// TODO: set cache headers
		if _, err := w.Write(sitemapXML); err != nil {
			golog.Errorf(err.Error())
		}
	})
}
