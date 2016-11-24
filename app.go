package main

import (
	"net/http"
	"os"

	"fmt"
	"strconv"
	"time"

	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/jmcvetta/neoism"
	"github.com/rcrowley/go-metrics"
)

func main() {
	log.Infof("Application starting with args %s", os.Args)
	app := cli.App("public-six-degrees", "A public API for accessing a set of endpoints serving information about peoples relationships to our content")

	neoURL := app.String(cli.StringOpt{
		Name:   "neo-url",
		Value:  "http://localhost:7474/db/data",
		Desc:   "neo4j endpoint URL",
		EnvVar: "NEO_URL",
	})
	port := app.String(cli.StringOpt{
		Name:   "port",
		Value:  "8080",
		Desc:   "Port to listen on",
		EnvVar: "APP_PORT",
	})
	graphiteTCPAddress := app.String(cli.StringOpt{
		Name:   "graphiteTCPAddress",
		Value:  "",
		Desc:   "Graphite TCP address, e.g. graphite.ft.com:2003. Leave as default if you do NOT want to output to graphite (e.g. if running locally)",
		EnvVar: "GRAPHITE_ADDRESS",
	})
	graphitePrefix := app.String(cli.StringOpt{
		Name:   "graphitePrefix",
		Value:  "",
		Desc:   "Prefix to use. Should start with content, include the environment, and the host name. e.g. content.test.public.people.api.ftaps59382-law1a-eu-t",
		EnvVar: "GRAPHITE_PREFIX",
	})
	logMetrics := app.Bool(cli.BoolOpt{
		Name:   "logMetrics",
		Value:  false,
		Desc:   "Whether to log metrics. Set to true if running locally and you want metrics output",
		EnvVar: "LOG_METRICS",
	})
	cacheDuration := app.String(cli.StringOpt{
		Name:   "cache-duration",
		Value:  "1h",
		Desc:   "Duration Get requests should be cached for. e.g. 2h45m would set the max-age value to '7440' seconds",
		EnvVar: "CACHE_DURATION",
	})

	app.Action = func() {
		baseftrwapp.OutputMetricsIfRequired(*graphiteTCPAddress, *graphitePrefix, *logMetrics)

		log.Infof("public-six-degrees will listen on port: %s, connecting to: %s", *port, *neoURL)
		runServer(*neoURL, *port, *cacheDuration)
	}
	log.SetLevel(log.InfoLevel)
	log.Infof("Application started with args %s", os.Args)
	app.Run(os.Args)
}

func runServer(neoURL string, port string, cacheDuration string) {
	if duration, durationErr := time.ParseDuration(cacheDuration); durationErr != nil {
		log.Fatalf("Failed to parse cache duration string, %v", durationErr)
	} else {
		CacheControlHeader = fmt.Sprintf("max-age=%s, public", strconv.FormatFloat(duration.Seconds(), 'f', 0, 64))
	}

	db, err := neoism.Connect(neoURL)
	if err != nil {
		log.Fatalf("Error connecting to neo4j %s", err)
	}
	db.Session.Client = &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: 100}}

	SixDegreesDriver = NewCypherDriver(db, "local")

	servicesRouter := mux.NewRouter()

	// Healthchecks and standards first
	servicesRouter.HandleFunc("/__health", v1a.Handler("PublicSixDegrees Healthchecks",
		"Checks for accessing neo4j", HealthCheck()))

	// Then API specific ones: TODO
	servicesRouter.HandleFunc("/sixdegrees/connectedPeople", GetConnectedPeople).Methods("GET")
	servicesRouter.HandleFunc("/sixdegrees/mostMentionedPeople", GetMostMentionedPeople).Methods("GET")

	//servicesRouter.HandleFunc("/people/{uuid}", people.MethodNotAllowedHandler)

	var monitoringRouter http.Handler = servicesRouter
	monitoringRouter = httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), monitoringRouter)
	monitoringRouter = httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry, monitoringRouter)

	// The following endpoints should not be monitored or logged (varnish calls one of these every second, depending on config)
	// The top one of these build info endpoints feels more correct, but the lower one matches what we have in Dropwizard,
	// so it's what apps expect currently same as ping, the content of build-info needs more definition
	http.HandleFunc(status.PingPath, status.PingHandler)
	http.HandleFunc(status.PingPathDW, status.PingHandler)
	http.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)
	http.HandleFunc(status.BuildInfoPathDW, status.BuildInfoHandler)
	http.HandleFunc("/__gtg", GoodToGo)
	http.Handle("/", monitoringRouter)

	if err := http.ListenAndServe(":"+port, handlers.CORS()(monitoringRouter)); err != nil {
		log.Fatalf("Unable to start server: %v", err)
	}
}
