package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"
	"github.com/Financial-Times/go-fthealth/v1a"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"time"
)

const knownUUID = "12345"

type test struct {
	name                       string
	req                        *http.Request
	driver                     *dummyDriver
	statusCode                 int
	contentType                string // Contents of the Content-Type header
	body                       string
	expectedResultLimit        int
	expectedFromDateEpoch      int64
	expectedToDateEpoch        int64
	expectedMinimumConnections int
	expectedContentLimit       int
}

func TestGetConnectedPeople(t *testing.T) {
	assert := assert.New(t)
	tests := []test{
		{
			name:                       "Success",
			req:                        newRequest("GET", fmt.Sprintf("/sixdegrees/connectedPeople?uuid=%s", knownUUID), "application/json", nil),
			driver:                     &dummyDriver{contentUUID: knownUUID},
			statusCode:                 http.StatusOK,
			contentType:                "",
			body:                       "[]",
			expectedResultLimit:        defaultConnectedPeopleResultLimit,
			expectedFromDateEpoch:      getDefaultFromDate().Unix(),
			expectedToDateEpoch:        getDefaultToDate().Unix(),
			expectedMinimumConnections: defaultMinConnections,
			expectedContentLimit:       defaultContentLimit,
		},
		{
			name:                       "SuccessWithResultLimitAndMinimumConnectionsAndContentLimit",
			req:                        newRequest("GET", fmt.Sprintf("/sixdegrees/connectedPeople?uuid=%s&limit=5&minimumConnections=2&contentLimit=10", knownUUID), "application/json", nil),
			driver:                     &dummyDriver{contentUUID: knownUUID},
			statusCode:                 http.StatusOK,
			contentType:                "",
			body:                       "[]",
			expectedResultLimit:        5,
			expectedFromDateEpoch:      getDefaultFromDate().Unix(),
			expectedToDateEpoch:        getDefaultToDate().Unix(),
			expectedMinimumConnections: 2,
			expectedContentLimit:       10,
		},
		{
			name:                       "SuccessWithFromAndToDateProvidedWithinOneYearRange",
			req:                        newRequest("GET", fmt.Sprintf("/sixdegrees/connectedPeople?uuid=%s&fromDate=%s&toDate=%s", knownUUID, time.Now().AddDate(0, -5, 1).Format("2006-01-02"), time.Now().AddDate(0, 1, 1).Format("2006-01-02")), "application/json", nil),
			driver:                     &dummyDriver{contentUUID: knownUUID},
			statusCode:                 http.StatusOK,
			contentType:                "",
			body:                       "[]",
			expectedResultLimit:        defaultConnectedPeopleResultLimit,
			expectedFromDateEpoch:      time.Now().AddDate(0, -5, 1).Unix(),
			expectedToDateEpoch:        time.Now().AddDate(0, 1, 1).Unix(),
			expectedMinimumConnections: defaultMinConnections,
			expectedContentLimit:       defaultContentLimit,
		},
		{
			name:                       "SuccessWithFromAndToDateProvidedWithYearRangeOutsidePermitted",
			req:                        newRequest("GET", fmt.Sprintf("/sixdegrees/connectedPeople?uuid=%s&fromDate=%s&toDate=%s", knownUUID, time.Now().AddDate(-2, 0, 1).Format("2006-01-02"), time.Now().AddDate(0, 1, 1).Format("2006-01-02")), "application/json", nil),
			driver:                     &dummyDriver{contentUUID: knownUUID},
			statusCode:                 http.StatusOK,
			contentType:                "",
			body:                       "[]",
			expectedResultLimit:        defaultConnectedPeopleResultLimit,
			expectedFromDateEpoch:      time.Now().AddDate(-2, 0, 1).Unix(),
			expectedToDateEpoch:        time.Now().AddDate(-1, 0, 1).Unix(),
			expectedMinimumConnections: defaultMinConnections,
			expectedContentLimit:       defaultContentLimit,
		},
		{
			name:                       "SuccessWithFromDateLaterThanToDate",
			req:                        newRequest("GET", fmt.Sprintf("/sixdegrees/connectedPeople?uuid=%s&fromDate=%s&toDate=%s", knownUUID, time.Now().AddDate(0, 0, 1).Format("2006-01-02"), time.Now().AddDate(0, 0, -1).Format("2006-01-02")), "application/json", nil),
			driver:                     &dummyDriver{contentUUID: knownUUID},
			statusCode:                 http.StatusOK,
			contentType:                "",
			body:                       "[]",
			expectedResultLimit:        defaultConnectedPeopleResultLimit,
			expectedFromDateEpoch:      time.Now().AddDate(0, 0, -8).Unix(),
			expectedToDateEpoch:        time.Now().AddDate(0, 0, -1).Unix(),
			expectedMinimumConnections: defaultMinConnections,
			expectedContentLimit:       defaultContentLimit,
		},
		{
			name:                       "FailureWithInvalidFromDate",
			req:                        newRequest("GET", fmt.Sprintf("/sixdegrees/connectedPeople?uuid=%s&fromDate=FAIL", knownUUID), "application/json", nil),
			driver:                     &dummyDriver{contentUUID: knownUUID},
			statusCode:                 http.StatusBadRequest,
			contentType:                "",
			body:                       message("Error converting toDate or fromDate query params: fromDate=FAIL, toDate="),
			expectedResultLimit:        0,
			expectedFromDateEpoch:      0,
			expectedToDateEpoch:        0,
			expectedMinimumConnections: 0,
			expectedContentLimit:       0,
		},
		{
			name:                       "FailureWithInvalidToDate",
			req:                        newRequest("GET", fmt.Sprintf("/sixdegrees/connectedPeople?uuid=%s&toDate=FAIL", knownUUID), "application/json", nil),
			driver:                     &dummyDriver{contentUUID: knownUUID},
			statusCode:                 http.StatusBadRequest,
			contentType:                "",
			body:                       message("Error converting toDate or fromDate query params: fromDate=, toDate=FAIL"),
			expectedResultLimit:        0,
			expectedFromDateEpoch:      0,
			expectedToDateEpoch:        0,
			expectedMinimumConnections: 0,
			expectedContentLimit:       0,
		},
		{
			name:                       "FailureWithInvalidResultLimit",
			req:                        newRequest("GET", fmt.Sprintf("/sixdegrees/connectedPeople?uuid=%s&limit=FAIL", knownUUID), "application/json", nil),
			driver:                     &dummyDriver{contentUUID: knownUUID},
			statusCode:                 http.StatusBadRequest,
			contentType:                "",
			body:                       message("Error converting limit query param, err=strconv.ParseInt: parsing \\\"FAIL\\\": invalid syntax"),
			expectedResultLimit:        0,
			expectedFromDateEpoch:      0,
			expectedToDateEpoch:        0,
			expectedMinimumConnections: 0,
			expectedContentLimit:       0,
		},
		{
			name:                       "FailureWithInvalidContentLimit",
			req:                        newRequest("GET", fmt.Sprintf("/sixdegrees/connectedPeople?uuid=%s&contentLimit=FAIL", knownUUID), "application/json", nil),
			driver:                     &dummyDriver{contentUUID: knownUUID},
			statusCode:                 http.StatusBadRequest,
			contentType:                "",
			body:                       message("Error converting contentLimit query param, err=strconv.ParseInt: parsing \\\"FAIL\\\": invalid syntax"),
			expectedResultLimit:        0,
			expectedFromDateEpoch:      0,
			expectedToDateEpoch:        0,
			expectedMinimumConnections: 0,
			expectedContentLimit:       0,
		},
		{
			name:                       "FailureWithInvalidMinConnections",
			req:                        newRequest("GET", fmt.Sprintf("/sixdegrees/connectedPeople?uuid=%s&minimumConnections=FAIL", knownUUID), "application/json", nil),
			driver:                     &dummyDriver{contentUUID: knownUUID},
			statusCode:                 http.StatusBadRequest,
			contentType:                "",
			body:                       message("Error converting minimumConnections query param, err=strconv.ParseInt: parsing \\\"FAIL\\\": invalid syntax"),
			expectedResultLimit:        0,
			expectedFromDateEpoch:      0,
			expectedToDateEpoch:        0,
			expectedMinimumConnections: 0,
			expectedContentLimit:       0,
		},
		{
			name:                       "NotFound",
			req:                        newRequest("GET", fmt.Sprintf("/sixdegrees/connectedPeople?uuid=%s", "99999"), "application/json", nil),
			driver:                     &dummyDriver{contentUUID: knownUUID},
			statusCode:                 http.StatusNotFound,
			contentType:                "",
			body:                       message("No connected people found for person with uuid 99999"),
			expectedResultLimit:        defaultConnectedPeopleResultLimit,
			expectedFromDateEpoch:      time.Now().AddDate(0, 0, -7).Unix(),
			expectedToDateEpoch:        time.Now().Unix(),
			expectedMinimumConnections: defaultMinConnections,
			expectedContentLimit:       defaultContentLimit,
		},
		{
			name:                       "ReadError",
			req:                        newRequest("GET", fmt.Sprintf("/sixdegrees/connectedPeople?uuid=%s", knownUUID), "application/json", nil),
			driver:                     &dummyDriver{contentUUID: knownUUID, shouldFail: true},
			statusCode:                 http.StatusInternalServerError,
			contentType:                "",
			body:                       message("Error retrieving result for 12345, err=TEST failing to READ"),
			expectedResultLimit:        defaultConnectedPeopleResultLimit,
			expectedFromDateEpoch:      time.Now().AddDate(0, 0, -7).Unix(),
			expectedToDateEpoch:        time.Now().Unix(),
			expectedMinimumConnections: defaultMinConnections,
			expectedContentLimit:       defaultContentLimit,
		},
	}

	for _, test := range tests {
		rec := httptest.NewRecorder()
		router(httpHandlers{test.driver, "max-age=360, public"}).ServeHTTP(rec, test.req)
		assert.True(test.statusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.statusCode))
		assert.JSONEq(test.body, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
		assert.Equal(test.expectedResultLimit, test.driver.argLimit, fmt.Sprintf("%s: Wrong limit", test.name))
		assert.Equal(time.Unix(test.expectedFromDateEpoch, 0).Format("2006-01-02"), time.Unix(test.driver.argFromDateEpoch, 0).Format("2006-01-02"), fmt.Sprintf("%s: Wrong from date", test.name))
		assert.Equal(time.Unix(test.expectedToDateEpoch, 0).Format("2006-01-02"), time.Unix(test.driver.argToDateEpoch, 0).Format("2006-01-02"), fmt.Sprintf("%s: Wrong to date", test.name))
		assert.Equal(test.expectedMinimumConnections, test.driver.argMinimumConnections, fmt.Sprintf("%s: Wrong minimum connections", test.name))
		assert.Equal(test.expectedContentLimit, test.driver.argContentLimit, fmt.Sprintf("%s: Wrong content limit", test.name))
	}
}

func TestGetMostMentionedPeople(t *testing.T) {
	assert := assert.New(t)
	tests := []test{
		{
			name:                  "Success",
			req:                   newRequest("GET", "/sixdegrees/mostMentionedPeople", "application/json", nil),
			driver:                &dummyDriver{},
			statusCode:            http.StatusOK,
			contentType:           "",
			body:                  "[]",
			expectedResultLimit:   defaultMostMentionedPeopleResultLimit,
			expectedFromDateEpoch: getDefaultFromDate().Unix(),
			expectedToDateEpoch:   getDefaultToDate().Unix(),
		},
		{
			name:                  "SuccessWithResultLimit",
			req:                   newRequest("GET", "/sixdegrees/mostMentionedPeople?limit=5", "application/json", nil),
			driver:                &dummyDriver{},
			statusCode:            http.StatusOK,
			contentType:           "",
			body:                  "[]",
			expectedResultLimit:   5,
			expectedFromDateEpoch: getDefaultFromDate().Unix(),
			expectedToDateEpoch:   getDefaultToDate().Unix(),
		},
		{
			name:                  "SuccessWithFromAndToDateProvidedWithinOneYearRange",
			req:                   newRequest("GET", fmt.Sprintf("/sixdegrees/mostMentionedPeople?fromDate=%s&toDate=%s", time.Now().AddDate(0, -5, 1).Format("2006-01-02"), time.Now().AddDate(0, 1, 1).Format("2006-01-02")), "application/json", nil),
			driver:                &dummyDriver{},
			statusCode:            http.StatusOK,
			contentType:           "",
			body:                  "[]",
			expectedResultLimit:   defaultMostMentionedPeopleResultLimit,
			expectedFromDateEpoch: time.Now().AddDate(0, -5, 1).Unix(),
			expectedToDateEpoch:   time.Now().AddDate(0, 1, 1).Unix(),
		},
		{
			name:                  "SuccessWithFromAndToDateProvidedWithYearRangeOutsidePermitted",
			req:                   newRequest("GET", fmt.Sprintf("/sixdegrees/mostMentionedPeople?fromDate=%s&toDate=%s", time.Now().AddDate(-2, 0, 1).Format("2006-01-02"), time.Now().AddDate(0, 1, 1).Format("2006-01-02")), "application/json", nil),
			driver:                &dummyDriver{},
			statusCode:            http.StatusOK,
			contentType:           "",
			body:                  "[]",
			expectedResultLimit:   defaultMostMentionedPeopleResultLimit,
			expectedFromDateEpoch: time.Now().AddDate(-2, 0, 1).Unix(),
			expectedToDateEpoch:   time.Now().AddDate(-1, 0, 1).Unix(),
		},
		{
			name:                  "SuccessWithFromDateLaterThanToDate",
			req:                   newRequest("GET", fmt.Sprintf("/sixdegrees/mostMentionedPeople?&fromDate=%s&toDate=%s", time.Now().AddDate(0, 0, 1).Format("2006-01-02"), time.Now().AddDate(0, 0, -1).Format("2006-01-02")), "application/json", nil),
			driver:                &dummyDriver{},
			statusCode:            http.StatusOK,
			contentType:           "",
			body:                  "[]",
			expectedResultLimit:   defaultMostMentionedPeopleResultLimit,
			expectedFromDateEpoch: time.Now().AddDate(0, 0, -8).Unix(),
			expectedToDateEpoch:   time.Now().AddDate(0, 0, -1).Unix(),
		},
		{
			name:        "FailureWithInvalidFromDate",
			req:         newRequest("GET", "/sixdegrees/mostMentionedPeople?fromDate=FAIL", "application/json", nil),
			driver:      &dummyDriver{},
			statusCode:  http.StatusBadRequest,
			contentType: "",
			body:        message("Error converting toDate or fromDate query params: fromDate=FAIL, toDate="),
		},
		{
			name:        "FailureWithInvalidToDate",
			req:         newRequest("GET", "/sixdegrees/mostMentionedPeople?toDate=FAIL", "application/json", nil),
			driver:      &dummyDriver{},
			statusCode:  http.StatusBadRequest,
			contentType: "",
			body:        message("Error converting toDate or fromDate query params: fromDate=, toDate=FAIL"),
		},
		{
			name:        "FailureWithInvalidResultLimit",
			req:         newRequest("GET", "/sixdegrees/mostMentionedPeople?limit=FAIL", "application/json", nil),
			driver:      &dummyDriver{},
			statusCode:  http.StatusBadRequest,
			contentType: "",
			body:        message("Error converting limit query param, err=strconv.ParseInt: parsing \\\"FAIL\\\": invalid syntax"),
		},
		{
			name:                  "NotFound",
			req:                   newRequest("GET", "/sixdegrees/mostMentionedPeople", "application/json", nil),
			driver:                &dummyDriver{shouldReturnNotFound: true},
			statusCode:            http.StatusNotFound,
			contentType:           "",
			body:                  message("No result"),
			expectedResultLimit:   defaultMostMentionedPeopleResultLimit,
			expectedFromDateEpoch: time.Now().AddDate(0, 0, -7).Unix(),
			expectedToDateEpoch:   time.Now().Unix(),
		},
		{
			name:                  "ReadError",
			req:                   newRequest("GET", "/sixdegrees/mostMentionedPeople", "application/json", nil),
			driver:                &dummyDriver{shouldFail: true},
			statusCode:            http.StatusInternalServerError,
			contentType:           "",
			body:                  message("Error retrieving result from DB"),
			expectedResultLimit:   defaultMostMentionedPeopleResultLimit,
			expectedFromDateEpoch: time.Now().AddDate(0, 0, -7).Unix(),
			expectedToDateEpoch:   time.Now().Unix(),
		},
	}

	for _, test := range tests {
		rec := httptest.NewRecorder()
		router(httpHandlers{test.driver, "max-age=360, public"}).ServeHTTP(rec, test.req)
		assert.True(test.statusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.statusCode))
		assert.JSONEq(test.body, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
		assert.Equal(test.expectedResultLimit, test.driver.argLimit, fmt.Sprintf("%s: Wrong limit", test.name))
		assert.Equal(time.Unix(test.expectedFromDateEpoch, 0).Format("2006-01-02"), time.Unix(test.driver.argFromDateEpoch, 0).Format("2006-01-02"), fmt.Sprintf("%s: Wrong from date", test.name))
		assert.Equal(time.Unix(test.expectedToDateEpoch, 0).Format("2006-01-02"), time.Unix(test.driver.argToDateEpoch, 0).Format("2006-01-02"), fmt.Sprintf("%s: Wrong to date", test.name))
	}
}
func TestCheckConnectivity(t *testing.T) {
	assert := assert.New(t)
	tests := []test{
		{
			name:        "HealthSuccess",
			req:         newRequest("GET", "/__health", "application/json", nil),
			driver:      &dummyDriver{},
			statusCode:  http.StatusOK,
			contentType: "",
			body:        `{"checks":[{"businessImpact":"Unable to respond to Public Six Degrees","checkOutput":"Connectivity to neo4j is ok","lastUpdated":"2016-12-13T16:03:31.2382547+02:00","name":"Check connectivity to Neo4j - neoUrl is a parameter in hieradata for this service","ok":true,"panicGuide":"https://dewey.ft.com/public-six-degrees-api.html","severity":1,"technicalSummary":"Cannot connect to Neo4j. If this check fails, check that Neo4j instance is up and running. You can find\n\t\t\t\tthe neoUrl as a parameter in hieradata for this service. "}],"description":"Checks for accessing neo4j","name":"PublicSixDegrees Healthchecks","schemaVersion":1,"ok":true}`,
		},
		{
			name:        "HealthError",
			req:         newRequest("GET", "/__health", "application/json", nil),
			driver:      &dummyDriver{shouldFail: true},
			statusCode:  http.StatusOK,
			contentType: "",
			body:        `{"checks":[{"businessImpact":"Unable to respond to Public Six Degrees","checkOutput":"TEST failing check connectivity","lastUpdated":"2016-12-13T16:03:31.2387546+02:00","name":"Check connectivity to Neo4j - neoUrl is a parameter in hieradata for this service","ok":false,"panicGuide":"https://dewey.ft.com/public-six-degrees-api.html","severity":1,"technicalSummary":"Cannot connect to Neo4j. If this check fails, check that Neo4j instance is up and running. You can find\n\t\t\t\tthe neoUrl as a parameter in hieradata for this service. "}],"description":"Checks for accessing neo4j","name":"PublicSixDegrees Healthchecks","schemaVersion":1,"ok":false,"severity":1}`,
		},
	}

	for _, test := range tests {
		rec := httptest.NewRecorder()

		httpHandler := httpHandlers{test.driver, "max-age=360, public"}
		router := mux.NewRouter()
		router.HandleFunc("/__health", v1a.Handler("PublicSixDegrees Healthchecks",
			"Checks for accessing neo4j", httpHandler.HealthCheck()))
		router.ServeHTTP(rec, test.req)
		assert.True(test.statusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.statusCode))

		var actualCheckResult, expectedCheckResult v1a.CheckResult
		json.Unmarshal([]byte(test.body), &expectedCheckResult)
		err := json.Unmarshal([]byte(rec.Body.String()), &actualCheckResult)
		assert.NoError(err, fmt.Sprintf("%s: Parse error for body", test.name))

		assert.Equal(expectedCheckResult.Ack, actualCheckResult.Ack)
		assert.Equal(expectedCheckResult.BusinessImpact, actualCheckResult.BusinessImpact)
		assert.Equal(expectedCheckResult.Name, actualCheckResult.Name)
		assert.Equal(expectedCheckResult.Ok, actualCheckResult.Ok)
		assert.Equal(expectedCheckResult.Output, actualCheckResult.Output)
		assert.Equal(expectedCheckResult.PanicGuide, actualCheckResult.PanicGuide)
		assert.Equal(expectedCheckResult.Severity, actualCheckResult.Severity)
		assert.Equal(expectedCheckResult.TechnicalSummary, actualCheckResult.TechnicalSummary)
		assert.WithinDuration(expectedCheckResult.LastUpdated, actualCheckResult.LastUpdated, 3*time.Second)

	}
}

func newRequest(method, url, contentType string, body []byte) *http.Request {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", contentType)
	return req
}

func message(errMsg string) string {
	return fmt.Sprintf("{\"message\": \"%s\"}\n", errMsg)
}

type dummyDriver struct {
	contentUUID           string
	shouldFail            bool
	argLimit              int
	argFromDateEpoch      int64
	argToDateEpoch        int64
	argMinimumConnections int
	argContentLimit       int
	shouldReturnNotFound  bool
}

func (ds *dummyDriver) ConnectedPeople(uuid string, fromDateEpoch int64, toDateEpoch int64, limit int, minimumConnections int, contentLimit int) ([]ConnectedPerson, bool, error) {
	ds.captureConnectedPeopleArgs(fromDateEpoch, toDateEpoch, limit, minimumConnections, contentLimit)

	if ds.shouldFail {
		return nil, false, errors.New("TEST failing to READ")
	}
	if uuid == ds.contentUUID {
		return []ConnectedPerson{}, true, nil
	}
	return nil, false, nil
}

func (ds *dummyDriver) captureConnectedPeopleArgs(fromDateEpoch int64, toDateEpoch int64, limit int, minimumConnections int, contentLimit int) {
	ds.argFromDateEpoch = fromDateEpoch
	ds.argToDateEpoch = toDateEpoch
	ds.argLimit = limit
	ds.argMinimumConnections = minimumConnections
	ds.argContentLimit = contentLimit
}

func (ds *dummyDriver) MostMentioned(fromDateEpoch int64, toDateEpoch int64, limit int) ([]Thing, bool, error) {
	ds.captureMostMentionedPeopleArgs(fromDateEpoch, toDateEpoch, limit)

	if ds.shouldFail {
		return nil, false, errors.New("TEST failing to READ")
	}

	if ds.shouldReturnNotFound {
		return []Thing{}, false, nil
	}
	return []Thing{}, true, nil
}

func (ds *dummyDriver) captureMostMentionedPeopleArgs(fromDateEpoch int64, toDateEpoch int64, limit int) {
	ds.argFromDateEpoch = fromDateEpoch
	ds.argToDateEpoch = toDateEpoch
	ds.argLimit = limit
}

func (ds *dummyDriver) CheckConnectivity() error {
	if ds.shouldFail {
		return errors.New("TEST failing check connectivity")
	}
	return nil
}