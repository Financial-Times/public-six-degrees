package main

import (
	"fmt"
	"net/http"

	"encoding/json"
	"net/url"

	"github.com/Financial-Times/go-fthealth/v1a"

	"time"

	log "github.com/Sirupsen/logrus"
	"strconv"
)

// PeopleDriver for cypher queries
var SixDegreesDriver Driver
var CacheControlHeader string

//var maxAge = 24 * time.Hour

// HealthCheck does something
func HealthCheck() v1a.Check {
	return v1a.Check{
		BusinessImpact: "Unable to respond to Public Six Degree",
		Name:           "Check connectivity to Neo4j - neoUrl is a parameter in hieradata for this service",
		PanicGuide:     "https://sites.google.com/a/ft.com/ft-technology-service-transition/home/run-book-library/public-people-api",
		Severity:       1,
		TechnicalSummary: `Cannot connect to Neo4j. If this check fails, check that Neo4j instance is up and running. You can find
				the neoUrl as a parameter in hieradata for this service. `,
		Checker: Checker,
	}
}

// Checker does more stuff
func Checker() (string, error) {
	err := SixDegreesDriver.CheckConnectivity()
	if err == nil {
		return "Connectivity to neo4j is ok", err
	}
	return "Error connecting to neo4j", err
}

// Ping says pong
func Ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

//GoodToGo returns a 503 if the healthcheck fails - suitable for use from varnish to check availability of a node
func GoodToGo(writer http.ResponseWriter, req *http.Request) {
	if _, err := Checker(); err != nil {
		writer.WriteHeader(http.StatusServiceUnavailable)
	}

}

// BuildInfoHandler - This is a stop gap and will be added to when we can define what we should display here
func BuildInfoHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "build-info")
}

// MethodNotAllowedHandler handles 405
func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	return
}

// GetPerson is the public API
func GetMostMentionedPeople(w http.ResponseWriter, r *http.Request) {
	limitParam := r.URL.Query().Get("limit")
	fromDateParam := r.URL.Query().Get("fromDate")
	toDateParam := r.URL.Query().Get("toDate")

	var limit int
	var fromDate, toDate time.Time
	var err error

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	//  Defaulting most mentioned limit to 20
	if limitParam == "" {
		log.Infof("No limit supplied therefore defaulting to 20")
		limit = 20
	} else {
		limit, err = strconv.Atoi(limitParam)
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Defaulting to a week ago
	if fromDateParam == "" {
		log.Infof("No fromDate supplied therefore defaulting to week ago")
		fromDate = time.Now().AddDate(0, 0, -7)
	} else {
		fromDate, _ = convertAnnotatedDateToDateTime(fromDateParam)
	}

	// Defaulting to now
	if toDateParam == "" {
		log.Infof("No toDate supplied therefore defaulting to now")
		toDate = time.Now()
	} else {
		toDate, _ = convertAnnotatedDateToDateTime(toDateParam)
	}

	// Defaulting fromDate to a week before toDate if the toDate is earlier than fromDate
	if toDate.Before(fromDate) {
		log.Infof("toDate cannot be earlier than fromDate, defaulting fromDate to a week from toDate")
		fromDate = toDate.AddDate(0, 0, -7)
	}

	// Restrict query for 1 year period
	fromDatePlusAYear := fromDate.AddDate(1, 0, 0)
	if fromDatePlusAYear.Before(toDate) {
		log.Infof("The given time period is greater than a year. Defaulting to a year from the given from date")
		toDate = fromDatePlusAYear
	}

	log.Infof("The given period is from %v to %v\n", fromDate.String(), toDate.String())

	people, found, err := SixDegreesDriver.MostMentioned(fromDate.Unix(), toDate.Unix(), limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Nothing found."}`))
		return
	}

	w.Header().Set("Cache-Control", CacheControlHeader)
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(people); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// GetPerson is the public API
func GetConnectedPeople(w http.ResponseWriter, request *http.Request) {
	m, _ := url.ParseQuery(request.URL.RawQuery)

	minimumConnectionsParam := m.Get("minimumConnections")
	limitParam := m.Get("limit")
	fromDateParam := m.Get("fromDate")
	toDateParam := m.Get("toDate")
	contentLimitParam := m.Get("contentLimit")

	uuid := m.Get("uuid")

	if minimumConnectionsParam == "" {
		minimumConnectionsParam = "5"
	}

	if limitParam == "" {
		limitParam = "10"
	}

	if contentLimitParam == "" {
		contentLimitParam = "3"
		log.Infof("No contentLimit supplied, defaulting contentLimit to %s", contentLimitParam)
	}

	var fromDate,toDate time.Time

	// Defaulting to a week ago
	if fromDateParam == "" {
		log.Infof("No fromDate supplied therefore defaulting to week ago")
		fromDate = time.Now().AddDate(0, 0, -7)
	} else {
		fromDate, _ = convertAnnotatedDateToDateTime(fromDateParam)
	}

	if toDateParam == "" {
		log.Infof("No toDate supplied therefore defaulting to now")
		toDate = time.Now()
	} else {
		toDate, _ = convertAnnotatedDateToDateTime(toDateParam)
	}

	// Defaulting fromDate to a week before toDate if the toDate is earlier than fromDate
	if toDate.Before(fromDate) {
		log.Infof("toDate cannot be earlier than fromDate, defaulting fromDate to a week from toDate")
		fromDate = toDate.AddDate(0, 0, -7)
	}

	// Restrict query for 1 year period
	fromDatePlusAYear := fromDate.AddDate(1, 0, 0)
	if fromDatePlusAYear.Before(toDate) {
		log.Infof("The given time period is greater than a year. Defaulting to a year from the given from date")
		toDate = fromDatePlusAYear
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	log.Infof("The given period is from %v to %v\n", fromDate.String(), toDate.String())

	minimumConnections, err := strconv.Atoi(minimumConnectionsParam)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	limit, err := strconv.Atoi(limitParam)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	contentLimit, err := strconv.Atoi(contentLimitParam)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	connectedPeople, _, _ := SixDegreesDriver.ConnectedPeople(uuid, fromDate.Unix(), toDate.Unix(), limit, minimumConnections, contentLimit)

	//samplePerson1 := Thing{"id " + uuid, "apiurl", "Angela Merkel"}
	//samplePerson2 := Thing{"id " + uuid, "apiurl", "David Cameron"}
	//sampleConnectedPerson1 := ConnectedPerson{samplePerson1, 534}
	//sampleConnectedPerson2 := ConnectedPerson{samplePerson2, 54}
	//sampleConnectedPersonSlice := []ConnectedPerson{sampleConnectedPerson1, sampleConnectedPerson2}
	//sampleConnectedPeople := ConnectedPeople{sampleConnectedPersonSlice}
	w.Header().Set("Cache-Control", CacheControlHeader)
	w.WriteHeader(http.StatusOK)
	//w.Write([]byte(`{"message": "hello world", "uuid": "` + uuid + `"}`))
	//json.NewEncoder(w).Encode(sampleConnectedPersonSlice)
	json.NewEncoder(w).Encode(connectedPeople)
}

func convertAnnotatedDateToDateTime(annotatedDateString string) (time.Time, error) {
	datetime, err := time.Parse("2006-01-02", annotatedDateString)

	if err != nil {
		return time.Time{}, err
	}

	return datetime, nil
}
