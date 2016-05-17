package main

import (
	"fmt"
	"net/http"

	"github.com/Financial-Times/go-fthealth/v1a"
	"net/url"
	"encoding/json"
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
	//vars := mux.Vars(r)
	//
	//m, _ := url.ParseQuery(r.URL.RawQuery)
	//
	//_, limit := m["limit"]
	//_, fromDate := m["fromDate"]
	//_, toDate := m["toDate"]

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Defaulting most mentioned amount to 20
	//if limit == "" {
	//	limit = "20"
	//}

	//thing, found, err := SixDegreesDriver.MostMentioned(fromDate, toDate, limit)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	// TODO: Check this
	//	//w.Write([]byte(`{"message": "` + err.Error() + `"}`))
	//	return
	//}
	//if !found {
	//	w.WriteHeader(http.StatusNotFound)
	//	w.Write([]byte(`{"message":"Nothing found."}`))
	//	return
	//}

	w.Header().Set("Cache-Control", CacheControlHeader)
	w.WriteHeader(http.StatusOK)

	//if err = json.NewEncoder(w).Encode(thing); err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	w.Write([]byte(`{"message":"Person could not be retrieved, err=` + err.Error() + `"}`))
	//}
}

// GetPerson is the public API
func GetConnectedPeople(w http.ResponseWriter, request *http.Request) {

	//vars := mux.Vars(request)

	m, _ := url.ParseQuery(request.URL.RawQuery)

	minimumConnectionsParam := m.Get("minimumConnections")
	limitParam := m.Get("limit")
	//fromDateParam := m.Get("fromDate")
	//toDateParam := m.Get("toDate")
	mockParam := m.Get("mock")
	//uuid := vars["uuid"]
	uuid := m.Get("uuid")

	if minimumConnectionsParam == "" {
		minimumConnectionsParam = "5"
	}

	if limitParam == "" {
		limitParam = "10"
	}

	if mockParam == "" {
		mockParam = "false"
	}

	// TODO parse fromDate and toDate into data objects

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")


	//minimumConnections, err := strconv.ParseInt(minimumConnectionsParam, 10, 64)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	// TODO: Check this
	//	//w.Write([]byte(`{"message": "` + err.Error() + `"}`))
	//	return
	//}

	//limit, err := strconv.ParseInt(limitParam, 10, 64)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	// TODO: Check this
	//	//w.Write([]byte(`{"message": "` + err.Error() + `"}`))
	//	return
	//}

	//mock, err := strconv.ParseBool(mockParam)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	// TODO: Check this
	//	//w.Write([]byte(`{"message": "` + err.Error() + `"}`))
	//	return
	//}

	samplePerson1 := Thing{"id " + uuid, "apiurl", "Angela Merkel"}
	samplePerson2 := Thing{"id " + uuid, "apiurl", "David Cameron"}
	sampleConnectedPerson1 := ConnectedPerson{samplePerson1, 534}
	sampleConnectedPerson2 := ConnectedPerson{samplePerson2, 54}
	sampleConnectedPersonSlice := []ConnectedPerson{sampleConnectedPerson1, sampleConnectedPerson2}
	//sampleConnectedPeople := ConnectedPeople{sampleConnectedPersonSlice}
	w.Header().Set("Cache-Control", CacheControlHeader)
	w.WriteHeader(http.StatusOK)
	//w.Write([]byte(`{"message": "hello world", "uuid": "` + uuid + `"}`))
	json.NewEncoder(w).Encode(sampleConnectedPersonSlice)
}
