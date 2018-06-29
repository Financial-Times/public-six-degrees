package sixdegrees

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const knownUUID = "12345"

type handlerTestCase struct {
	name                       string
	req                        *http.Request
	driver                     *dummyDriver
	statusCode                 int
	contentType                string
	body                       string
	expectedResultLimit        int
	expectedFromDateEpoch      int64
	expectedToDateEpoch        int64
	expectedMinimumConnections int
	expectedContentLimit       int
}

func TestGetConnectedPeople(t *testing.T) {
	assert := assert.New(t)
	tests := []handlerTestCase{
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
		router := mux.NewRouter()
		handler := Handler{test.driver, "max-age=360, public"}
		handler.RegisterHandlers(router)
		router.ServeHTTP(rec, test.req)
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
	tests := []handlerTestCase{
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
		router := mux.NewRouter()
		handler := Handler{test.driver, "max-age=360, public"}
		handler.RegisterHandlers(router)
		router.ServeHTTP(rec, test.req)
		assert.True(test.statusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.statusCode))
		assert.JSONEq(test.body, rec.Body.String(), fmt.Sprintf("%s: Wrong body", test.name))
		assert.Equal(test.expectedResultLimit, test.driver.argLimit, fmt.Sprintf("%s: Wrong limit", test.name))
		assert.Equal(time.Unix(test.expectedFromDateEpoch, 0).Format("2006-01-02"), time.Unix(test.driver.argFromDateEpoch, 0).Format("2006-01-02"), fmt.Sprintf("%s: Wrong from date", test.name))
		assert.Equal(time.Unix(test.expectedToDateEpoch, 0).Format("2006-01-02"), time.Unix(test.driver.argToDateEpoch, 0).Format("2006-01-02"), fmt.Sprintf("%s: Wrong to date", test.name))
	}
}
func TestCheckConnectivity(t *testing.T) {
	assert := assert.New(t)

	timeTemplate := "2006-01-02T15:04:05.000Z"

	testSuccessTimeStr := "2016-12-13T16:03:31.2382547+02:00"
	testSuccessTime, _ := time.Parse(timeTemplate, testSuccessTimeStr)

	testSuccessResponse := fthealth.HealthResult{
		SchemaVersion: 1,
		SystemCode:    "public-six-degrees-api",
		Name:          "Public Six Degrees API",
		Description:   "Six Degrees Backend provides mostMentionedPeople and connectedPeople endpoints for Six Degrees Frontend.",
		Ok:            true,
		Checks: []fthealth.CheckResult{
			{
				Name:             "Check connectivity to Neo4j - neoUrl is a parameter in hieradata for this service",
				Ok:               true,
				Severity:         3,
				BusinessImpact:   "Unable to respond to Public Six Degrees",
				PanicGuide:       "https://dewey.ft.com/public-six-degrees-api.html",
				TechnicalSummary: "Cannot connect to Neo4j. If this check fails, check that Neo4j instance is up and running. You can find\n\t\t\t\tthe neoUrl as a parameter in hieradata for this service.",
				CheckOutput:      "Connectivity to neo4j is ok",
				LastUpdated:      testSuccessTime,
			},
		},
	}

	testErrorTimeStr := "2016-12-13T16:03:31.2387546+02:00"
	testErrorTime, _ := time.Parse(timeTemplate, testErrorTimeStr)

	testErrorResponse := fthealth.HealthResult{
		SchemaVersion: 1,
		SystemCode:    "public-six-degrees-api",
		Name:          "Public Six Degrees API",
		Description:   "Six Degrees Backend provides mostMentionedPeople and connectedPeople endpoints for Six Degrees Frontend.",
		Ok:            false,
		Severity:      3,
		Checks: []fthealth.CheckResult{
			{
				Name:             "Check connectivity to Neo4j - neoUrl is a parameter in hieradata for this service",
				Ok:               false,
				Severity:         3,
				BusinessImpact:   "Unable to respond to Public Six Degrees",
				PanicGuide:       "https://dewey.ft.com/public-six-degrees-api.html",
				TechnicalSummary: "Cannot connect to Neo4j. If this check fails, check that Neo4j instance is up and running. You can find\n\t\t\t\tthe neoUrl as a parameter in hieradata for this service.",
				CheckOutput:      "Error connecting to neo4j",
				LastUpdated:      testErrorTime,
			},
		},
	}

	testSuccessResponseStr, _ := json.Marshal(testSuccessResponse)
	testErrorResponseStr, _ := json.Marshal(testErrorResponse)

	tests := []handlerTestCase{
		{
			name:        "HealthSuccess",
			req:         newRequest("GET", "/__health", "application/json", nil),
			driver:      &dummyDriver{},
			statusCode:  http.StatusOK,
			contentType: "",
			body:        string(testSuccessResponseStr),
		},
		{
			name:        "HealthError",
			req:         newRequest("GET", "/__health", "application/json", nil),
			driver:      &dummyDriver{shouldFail: true},
			statusCode:  http.StatusOK,
			contentType: "",
			body:        string(testErrorResponseStr),
		},
	}

	for _, test := range tests {
		rec := httptest.NewRecorder()

		httpHandler := Handler{test.driver, "max-age=360, public"}
		router := mux.NewRouter()

		timedHC := fthealth.TimedHealthCheck{
			HealthCheck: fthealth.HealthCheck{
				SystemCode:  "public-six-degrees-api",
				Name:        "Public Six Degrees API",
				Description: "Six Degrees Backend provides mostMentionedPeople and connectedPeople endpoints for Six Degrees Frontend.",
				Checks:      []fthealth.Check{httpHandler.HealthCheck()},
			},
			Timeout: 10 * time.Second,
		}
		router.HandleFunc("/__health", fthealth.Handler(timedHC))
		router.ServeHTTP(rec, test.req)
		assert.True(test.statusCode == rec.Code, fmt.Sprintf("%s: Wrong response code, was %d, should be %d", test.name, rec.Code, test.statusCode))

		var actualHealthResult, expectedHealthResult fthealth.HealthResult
		json.Unmarshal([]byte(test.body), &expectedHealthResult)
		err := json.Unmarshal([]byte(rec.Body.String()), &actualHealthResult)
		assert.NoError(err, fmt.Sprintf("%s: Parse error for body", test.name))

		assert.Equal(expectedHealthResult.SchemaVersion, actualHealthResult.SchemaVersion)
		assert.Equal(expectedHealthResult.SystemCode, actualHealthResult.SystemCode)
		assert.Equal(expectedHealthResult.Name, actualHealthResult.Name)
		assert.Equal(expectedHealthResult.Description, actualHealthResult.Description)
		assert.Equal(expectedHealthResult.Ok, actualHealthResult.Ok)
		assert.Equal(expectedHealthResult.Severity, actualHealthResult.Severity)

		assert.Equal(expectedHealthResult.Checks[0].Ack, expectedHealthResult.Checks[0].Ack)
		assert.Equal(expectedHealthResult.Checks[0].BusinessImpact, expectedHealthResult.Checks[0].BusinessImpact)
		assert.Equal(expectedHealthResult.Checks[0].Name, expectedHealthResult.Checks[0].Name)
		assert.Equal(expectedHealthResult.Checks[0].Ok, expectedHealthResult.Checks[0].Ok)
		assert.Equal(expectedHealthResult.Checks[0].CheckOutput, expectedHealthResult.Checks[0].CheckOutput)
		assert.Equal(expectedHealthResult.Checks[0].PanicGuide, expectedHealthResult.Checks[0].PanicGuide)
		assert.Equal(expectedHealthResult.Checks[0].Severity, expectedHealthResult.Checks[0].Severity)
		assert.Equal(expectedHealthResult.Checks[0].TechnicalSummary, expectedHealthResult.Checks[0].TechnicalSummary)
		assert.WithinDuration(expectedHealthResult.Checks[0].LastUpdated, expectedHealthResult.Checks[0].LastUpdated, 3*time.Second)
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
