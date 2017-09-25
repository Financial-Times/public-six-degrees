package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	annrw "github.com/Financial-Times/annotations-rw-neo4j/annotations"
	"github.com/Financial-Times/base-ft-rw-app-go/baseftrwapp"
	cnt "github.com/Financial-Times/content-rw-neo4j/content"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	person "github.com/Financial-Times/people-rw-neo4j/people"
	"github.com/jmcvetta/neoism"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	contentUUID             = "3fc9fe3e-af8c-4f7f-961a-e5065392bb31"
	content2UUID            = "a435b4ec-b207-4dce-ac0a-f8e7bbef310b"
	personSiobhanMordenUUID = "13a9d251-71db-467a-af2f-7e56a61c910a"
	personBorisJohnsonUUID  = "b30ec30e-83ca-4e4a-b82f-db6f7a0bb16d"
)

type cypherTestCase struct {
	name                              string
	conn                              interceptingCypherConn
	uuid                              string
	fromDateEpoch                     int64
	toDateEpoch                       int64
	makeConnectedPeopleAssertions     func(*testing.T, []ConnectedPerson, bool, error, string)
	makeMostMentionedPeopleAssertions func(*testing.T, []Thing, bool, error, string)
}

func TestConnectedPeople(t *testing.T) {
	db := getDatabaseConnection(t)

	//We want to make sure we have an empty DB before and after we run the tests
	cleanDB(db, t)
	defer cleanDB(db, t)

	peopleRW := person.NewCypherPeopleService(db)
	require.NoError(t, peopleRW.Initialise())
	writeJsonToService(peopleRW, fmt.Sprintf("./fixtures/Person-Siobhan_Morden-%s.json", personSiobhanMordenUUID), t)
	writeJsonToService(peopleRW, fmt.Sprintf("./fixtures/Person-Boris_Johnson-%s.json", personBorisJohnsonUUID), t)

	contentRW := cnt.NewCypherContentService(db)
	require.NoError(t, contentRW.Initialise())
	writeJsonToService(contentRW, fmt.Sprintf("./fixtures/Content-%s.json", contentUUID), t)
	writeJsonToService(contentRW, fmt.Sprintf("./fixtures/Content-%s.json", content2UUID), t)


	annotationsRW := annrw.NewCypherAnnotationsService(db)
	require.NoError(t, annotationsRW.Initialise())
	writeJSONToAnnotationsService(annotationsRW, contentUUID, fmt.Sprintf("./fixtures/Annotations-%s-v2.json", contentUUID), t)
	writeJSONToAnnotationsService(annotationsRW, content2UUID, fmt.Sprintf("./fixtures/Annotations-%s-v2.json", content2UUID), t)

	tests := []cypherTestCase{
		{
			name: "SuccessWithConnectedPersonInTimeRange",
			conn: interceptingCypherConn{db: db},
			uuid: personBorisJohnsonUUID,
			makeConnectedPeopleAssertions: func(t *testing.T, connectedPeople []ConnectedPerson, found bool, err error, testName string) {
				assert.NoError(t, err, fmt.Sprintf("%s: Error found", testName))
				assert.True(t, found, fmt.Sprintf("%s: No result found", testName))
				assert.Equal(t, getExpectedConnectedPeople(), connectedPeople, fmt.Sprintf("%s: Actual connected people are different than expected", testName))
			},
			fromDateEpoch: getTimeEpoch("2016-12-12"),
			toDateEpoch:   getTimeEpoch("2016-12-16"),
		},
		{
			name: "SuccessWithoutConnectedPersonInTimeRange",
			conn: interceptingCypherConn{db: db},
			uuid: personBorisJohnsonUUID,
			makeConnectedPeopleAssertions: func(t *testing.T, connectedPeople []ConnectedPerson, found bool, err error, testName string) {
				assert.NoError(t, err, fmt.Sprintf("%s: Error found", testName))
				assert.False(t, found, fmt.Sprintf("%s: Result was found", testName))
				assert.Equal(t, []ConnectedPerson{}, connectedPeople, fmt.Sprintf("%s: Actual connected people are different than expected", testName))
			},
			fromDateEpoch: getTimeEpoch("2015-12-12"),
			toDateEpoch:   getTimeEpoch("2015-12-16"),
		},
		{
			name: "Failure",
			conn: interceptingCypherConn{db: db, shouldFail: true},
			uuid: personBorisJohnsonUUID,
			makeConnectedPeopleAssertions: func(t *testing.T, connectedPeople []ConnectedPerson, found bool, err error, testName string) {
				assert.Error(t, err, fmt.Sprintf("%s: Error not found", testName))
				assert.False(t, found, fmt.Sprintf("%s: Result was found", testName))
				assert.Equal(t, []ConnectedPerson{}, connectedPeople, fmt.Sprintf("%s: Actual connected people are different than expected", testName))
			},
			fromDateEpoch: getTimeEpoch("2015-12-12"),
			toDateEpoch:   getTimeEpoch("2015-12-16"),
		},
	}

	for _, test := range tests {
		connectedPeople, found, err := cypherDriver{test.conn}.ConnectedPeople(test.uuid, test.fromDateEpoch, test.toDateEpoch, 1, 1, 5)
		test.makeConnectedPeopleAssertions(t, connectedPeople, found, err, test.name)
	}
}

func TestMostMentionedPeople(t *testing.T) {
	db := getDatabaseConnection(t)

	//We want to make sure we have an empty DB before and after we run the tests
	cleanDB(db, t)
	defer cleanDB(db, t)

	peopleRW := person.NewCypherPeopleService(db)
	require.NoError(t, peopleRW.Initialise())
	writeJsonToService(peopleRW, fmt.Sprintf("./fixtures/Person-Siobhan_Morden-%s.json", personSiobhanMordenUUID), t)
	writeJsonToService(peopleRW, fmt.Sprintf("./fixtures/Person-Boris_Johnson-%s.json", personBorisJohnsonUUID), t)

	contentRW := cnt.NewCypherContentService(db)
	require.NoError(t, contentRW.Initialise())
	writeJsonToService(contentRW, fmt.Sprintf("./fixtures/Content-%s.json", contentUUID), t)
	writeJsonToService(contentRW, fmt.Sprintf("./fixtures/Content-%s.json", content2UUID), t)

	annotationsRW := annrw.NewCypherAnnotationsService(db)
	require.NoError(t, annotationsRW.Initialise())
	writeJSONToAnnotationsService(annotationsRW, contentUUID, fmt.Sprintf("./fixtures/Annotations-%s-v2.json", contentUUID), t)
	writeJSONToAnnotationsService(annotationsRW, content2UUID, fmt.Sprintf("./fixtures/Annotations-%s-v2.json", content2UUID), t)

	tests := []cypherTestCase{
		{
			name: "SuccessWithMostMentionedInTimeRange",
			conn: interceptingCypherConn{db: db},
			makeMostMentionedPeopleAssertions: func(t *testing.T, mentionedPeople []Thing, found bool, err error, testName string) {
				assert.NoError(t, err, fmt.Sprintf("%s: Error found", testName))
				assert.True(t, found, fmt.Sprintf("%s: No result found", testName))
				assert.Equal(t, getExpectedMostMentionedPeople(), mentionedPeople, fmt.Sprintf("%s: Actual most mentioned people are different than expected", testName))
			},
			fromDateEpoch: getTimeEpoch("2016-12-12"),
			toDateEpoch:   getTimeEpoch("2016-12-16"),
		},
		{
			name: "SuccessWithoutConnectedPersonInTimeRange",
			conn: interceptingCypherConn{db: db},
			makeMostMentionedPeopleAssertions: func(t *testing.T, mentionedPeople []Thing, found bool, err error, testName string) {
				assert.NoError(t, err, fmt.Sprintf("%s: Error found", testName))
				assert.False(t, found, fmt.Sprintf("%s: Result was found", testName))
				assert.Equal(t, []Thing{}, mentionedPeople, fmt.Sprintf("%s: Actual most mentioned people are different than expected", testName))
			},
			fromDateEpoch: getTimeEpoch("2015-12-12"),
			toDateEpoch:   getTimeEpoch("2015-12-16"),
		},
		{
			name: "Failure",
			conn: interceptingCypherConn{db: db, shouldFail: true},
			makeMostMentionedPeopleAssertions: func(t *testing.T, mentionedPeople []Thing, found bool, err error, testName string) {
				assert.Error(t, err, fmt.Sprintf("%s: Error not found", testName))
				assert.False(t, found, fmt.Sprintf("%s: Result was found", testName))
				assert.Equal(t, []Thing{}, mentionedPeople, fmt.Sprintf("%s: Actual most mentioned people are different than expected", testName))
			},
			fromDateEpoch: getTimeEpoch("2015-12-12"),
			toDateEpoch:   getTimeEpoch("2015-12-16"),
		},
	}

	for _, test := range tests {
		thingList, found, err := cypherDriver{test.conn}.MostMentioned(test.fromDateEpoch, test.toDateEpoch, 5)
		test.makeMostMentionedPeopleAssertions(t, thingList, found, err, test.name)
	}
}

type interceptingCypherConn struct {
	db         neoutils.NeoConnection
	shouldFail bool
}

func (c interceptingCypherConn) CypherBatch(cypher []*neoism.CypherQuery) error {
	if c.shouldFail {
		return fmt.Errorf("BOOM!")
	}
	return c.db.CypherBatch(cypher)
}

func (c interceptingCypherConn) EnsureConstraints(constraints map[string]string) error {
	return c.db.EnsureConstraints(constraints)
}

func (c interceptingCypherConn) EnsureIndexes(indexes map[string]string) error {
	return c.db.EnsureIndexes(indexes)
}

func writeJsonToService(service baseftrwapp.Service, pathToJsonFile string, t *testing.T) {
	f, err := os.Open(pathToJsonFile)
	assert.NoError(t, err)
	dec := json.NewDecoder(f)
	inst, _, err := service.DecodeJSON(dec)
	assert.NoError(t, err)
	err = service.Write(inst, "trans_id")
	require.NoError(t, err)
}

func getExpectedConnectedPeople() []ConnectedPerson {
	return []ConnectedPerson{
		{
			Person: Thing{
				ID:        "http://api.ft.com/things/13a9d251-71db-467a-af2f-7e56a61c910a",
				APIURL:    "http://api.ft.com/people/13a9d251-71db-467a-af2f-7e56a61c910a",
				PrefLabel: "Siobhan Morden",
			},
			Count: 2,
			Content: []Content{
				{
					ID:     "3fc9fe3e-af8c-4f7f-961a-e5065392bb31",
					APIURL: "http://api.ft.com/content/3fc9fe3e-af8c-4f7f-961a-e5065392bb31",
					Title:  "Bitcoin story makes Newsweek the headline",
				},
				{
					ID:     "a435b4ec-b207-4dce-ac0a-f8e7bbef310b",
					APIURL: "http://api.ft.com/content/a435b4ec-b207-4dce-ac0a-f8e7bbef310b",
					Title:  "Learn Golang",
				},
			},
		},
	}
}

func getExpectedMostMentionedPeople() []Thing {
	return []Thing{
		{
			ID:        fmt.Sprintf("http://api.ft.com/things/%s", personBorisJohnsonUUID),
			PrefLabel: "Boris Johnson",
		},
		{
			ID:        fmt.Sprintf("http://api.ft.com/things/%s", personSiobhanMordenUUID),
			PrefLabel: "Siobhan Morden",
		},
	}
}

func getTimeEpoch(date string) int64 {
	t, _ := time.Parse("2006-01-02", date)
	return t.Unix()
}

func writeJSONToAnnotationsService(service annrw.Service, contentUUID string, pathToJSONFile string, t *testing.T) {
	f, err := os.Open(pathToJSONFile)
	assert.NoError(t, err)
	dec := json.NewDecoder(f)
	inst, err := service.DecodeJSON(dec)
	assert.NoError(t, err, "Error parsing file %s", pathToJSONFile)
	err = service.Write(contentUUID, "v2", "", "trans_id", inst)
	assert.NoError(t, err)
}

func getDatabaseConnection(t *testing.T) neoutils.NeoConnection {
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "http://localhost:7474/db/data"
	}
	conf := neoutils.DefaultConnectionConfig()
	conf.Transactional = false
	db, err := neoutils.Connect(url, conf)
	require.NoError(t, err, "Failed to connect to Neo4j")
	return db
}

//DELETES ALL DATA! DO NOT USE IN PRODUCTION!!!
func cleanDB(db neoutils.NeoConnection, t *testing.T) {
	qs := []*neoism.CypherQuery{
		{
			Statement: "MATCH (a) DETACH DELETE a",
		},
	}
	err := db.CypherBatch(qs)
	assert.NoError(t, err)
}
