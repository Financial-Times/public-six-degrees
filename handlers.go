package main

import (
	"net/http"

	"encoding/json"
	"net/url"

	"github.com/Financial-Times/go-fthealth/v1a"

	"time"

	log "github.com/Sirupsen/logrus"
	"strconv"
)

type defaultTimeGetter func() time.Time

type httpHandlers struct {
	sixDegreesDriver   driver
	cacheControlHeader string
}

func (hh *httpHandlers) HealthCheck() v1a.Check {
	return v1a.Check{
		BusinessImpact: "Unable to respond to Public Six Degrees",
		Name:           "Check connectivity to Neo4j - neoUrl is a parameter in hieradata for this service",
		PanicGuide:     "https://dewey.ft.com/public-six-degrees-api.html",
		Severity:       1,
		TechnicalSummary: `Cannot connect to Neo4j. If this check fails, check that Neo4j instance is up and running. You can find
				the neoUrl as a parameter in hieradata for this service. `,
		Checker: hh.Checker,
	}
}

func (hh *httpHandlers) Checker() (string, error) {
	err := hh.sixDegreesDriver.CheckConnectivity()
	if err == nil {
		return "Connectivity to neo4j is ok", err
	}
	return "Error connecting to neo4j", err
}

//GoodToGo returns a 503 if the healthcheck fails - suitable for use from varnish to check availability of a node
func (hh *httpHandlers) GoodToGo(writer http.ResponseWriter, req *http.Request) {
	if _, err := hh.Checker(); err != nil {
		writer.WriteHeader(http.StatusServiceUnavailable)
	}

}

func (hh *httpHandlers) GetMostMentionedPeople(w http.ResponseWriter, r *http.Request) {
	limitParam := r.URL.Query().Get("limit")
	fromDateParam := r.URL.Query().Get("fromDate")
	toDateParam := r.URL.Query().Get("toDate")

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	limit, err := getLimit(limitParam, 20)
	if err != nil {
		log.Errorf("ERROR - %v\n", err)
		http.Error(w, "Error converting minimumConnections query param", http.StatusBadRequest)
		return
	}

	fromDate, toDate, err := getDateTimePeriod(fromDateParam, toDateParam)
	if err != nil {
		log.Errorf("ERROR - %v\n", err)
		http.Error(w, "Error converting toDate or fromDate query params", http.StatusBadRequest)
		return
	}

	people, found, err := hh.sixDegreesDriver.MostMentioned(fromDate.Unix(), toDate.Unix(), limit)
	if err != nil {
		log.Errorf("ERROR - %v\n", err)
		http.Error(w, "Error retrieving result from DB", http.StatusInternalServerError)
		return
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Nothing found."}`))
		return
	}

	w.Header().Set("Cache-Control", hh.cacheControlHeader)
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(people); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (hh *httpHandlers) GetConnectedPeople(w http.ResponseWriter, request *http.Request) {
	m, _ := url.ParseQuery(request.URL.RawQuery)

	minimumConnectionsParam := m.Get("minimumConnections")
	resultLimitParam := m.Get("limit")
	fromDateParam := m.Get("fromDate")
	toDateParam := m.Get("toDate")
	contentLimitParam := m.Get("contentLimit")

	uuid := m.Get("uuid")

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	fromDate, toDate, err := getDateTimePeriod(fromDateParam, toDateParam)
	if err != nil {
		log.Errorf("ERROR - %v\n", err)
		http.Error(w, "Error converting toDate or fromDate query params", http.StatusBadRequest)
		return
	}

	minimumConnections, err := getLimit(minimumConnectionsParam, 5)
	if err != nil {
		log.Errorf("ERROR - %v\n", err)
		http.Error(w, "Error converting minimumConnections query param", http.StatusBadRequest)
		return
	}

	resultLimit, err := getLimit(resultLimitParam, 10)
	if err != nil {
		log.Errorf("ERROR - %v\n", err)
		http.Error(w, "Error converting limit query param", http.StatusBadRequest)
		return
	}

	contentLimit, err := getLimit(contentLimitParam, 3)
	if err != nil {
		log.Errorf("ERROR - %v\n", err)
		http.Error(w, "Error converting contentLimit query param", http.StatusBadRequest)
		return
	}

	connectedPeople, _, err := hh.sixDegreesDriver.ConnectedPeople(uuid, fromDate.Unix(), toDate.Unix(), resultLimit, minimumConnections, contentLimit)
	if err != nil {
		log.Errorf("ERROR - %v\n", err)
		http.Error(w, "Error retrieving result from DB", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", hh.cacheControlHeader)
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(connectedPeople)
}

func getDateTimePeriod(fromDateParam string, toDateParam string) (fromDate time.Time, toDate time.Time, err error) {
	fromDate, err = getDate(fromDateParam, getDefaultFromDate)
	if err != nil {
		return
	}

	toDate, err = getDate(toDateParam, getDefaultToDate)
	if err != nil {
		return
	}

	//toDate cannot be earlier than fromDate, defaulting fromDate to a week from toDate
	if toDate.Before(fromDate) {
		fromDate = toDate.AddDate(0, 0, -7)
	}

	// Restrict query for 1 year period
	fromDatePlusAYear := fromDate.AddDate(1, 0, 0)
	if fromDatePlusAYear.Before(toDate) {
		toDate = fromDatePlusAYear
	}

	log.Infof("The given period is from %v to %v\n", fromDate.String(), toDate.String())
	return
}

func getDate(dateParam string, getDefaultTime defaultTimeGetter) (time.Time, error) {
	if dateParam == "" {
		return getDefaultTime(), nil
	}
	return convertDateStringToDateTime(dateParam)
}

func getLimit(limitParam string, defaultLimit int) (int, error) {
	if limitParam == "" {
		return defaultLimit, nil
	}
	return strconv.Atoi(limitParam)
}

func convertDateStringToDateTime(dateString string) (time.Time, error) {
	datetime, err := time.Parse("2006-01-02", dateString)

	if err != nil {
		return time.Time{}, err
	}

	return datetime, nil
}

func getDefaultFromDate() time.Time {
	return time.Now().AddDate(0, 0, -7)
}

func getDefaultToDate() time.Time {
	return time.Now()
}
