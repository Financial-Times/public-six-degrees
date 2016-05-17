package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
)

// Driver interface
type Driver interface {
	ConnectedPeople(uuid string, fromDateEpoch int64, toDateEpoch int64, limit int, minimumConnections int) (connectedPeople[] ConnectedPerson, found bool, err error)
	MostMentioned(fromDateEpoch int64, toDateEpoch int64, limit int) (thingList People, found bool, err error)
	CheckConnectivity() error
}

// CypherDriver struct
type CypherDriver struct {
	db  *neoism.Database
	env string
}

//NewCypherDriver instantiate driver
func NewCypherDriver(db *neoism.Database, env string) CypherDriver {
	return CypherDriver{db, env}
}

// CheckConnectivity tests neo4j by running a simple cypher query
func (pcw CypherDriver) CheckConnectivity() error {
	results := []struct {
		ID int
	}{}
	query := &neoism.CypherQuery{
		Statement: "MATCH (x) RETURN ID(x) LIMIT 1",
		Result:    &results,
	}
	err := pcw.db.Cypher(query)
	log.Debugf("CheckConnectivity results:%+v  err: %+v", results, err)
	return err
}

type neoReadStruct struct {
}

//func (pcw CypherDriver) MostMentioned(fromDate string, toDate string, limit int64) (Thing, bool, error) {
//	return Thing{"bla", "bla", "bla"}, true, nil
//}

//MATCH (c:Content) where c.publishedDateEpoch < 1460371050 and c.publishedDateEpoch > 1457605324
//WITH c
//MATCH (p:Person{uuid:"9421d9ee-7e0f-3f7c-8adc-ded83fabdb92"})<-[:MENTIONS]-(c)-[:MENTIONS]->(p2:Person)
//WITH p, count(distinct(c)) as cm, p2
//WHERE cm > 5
//return cm, p2 limit 25
func (pcw CypherDriver) ConnectedPeople(uuid string, fromDateEpoch int64, toDateEpoch int64, limit int, minimumConnections int) (connectedPeople[] ConnectedPerson, found bool, err error) {
	connectedPeople = []ConnectedPerson{}
	results := []struct {
		Rs []neoReadStruct
	}{}
	statement := `
	MATCH (c:Content) where c.publishedDateEpoch < {toDate} and c.publishedDateEpoch > {fromDate}
	WITH c
	MATCH (p:Person{uuid:{uuid}})<-[:MENTIONS]-(c)-[:MENTIONS]->(p2:Person)
	WITH p, count(distinct(c)) as cm, p2
	WHERE cm > {minimumConnections}
	WITH p2.uuid as uuid
	RETURN name, uuid, mentions ORDER BY mentions DESC LIMIT {limit}`
	//thing := Thing{}
	query := &neoism.CypherQuery{
		Statement:  statement,
		Parameters: neoism.Props{
			"uuid": uuid,
			"fromDate": fromDateEpoch,
			"toDate": toDateEpoch,
			"minimumConnections": minimumConnections,
			"limit": limit,
		},
		Result:     &results,
	}
	err = pcw.db.Cypher(query)
	if err != nil {
		log.Errorf("Error finding %v most mentioned people between %v and %v with the following statement: %v  Error: %v", limit, fromDateEpoch, toDateEpoch, query.Statement, err)
		return []ConnectedPerson{}, false, fmt.Errorf("Error finding %v most mentioned people between %v and %v", limit, fromDateEpoch, toDateEpoch)
	}
	log.Debugf("CypherResult MostMentioned was (fromDate=%v, toDate=%v): %+v", limit, fromDateEpoch, toDateEpoch, results)
	if (len(results)) == 0 || len(results[0].Rs) == 0 {
		return []ConnectedPerson{}, false, nil
	}
	log.Debugf("Returning %v", connectedPeople)
	return connectedPeople, true, nil
}

func (pcw CypherDriver) MostMentioned(fromDateEpoch int64, toDateEpoch int64, limit int) (thingList People, found bool, err error) {
	people := People{}
	results := []struct {
		Rs []neoReadStruct
	}{}
	query := &neoism.CypherQuery{
		Statement:  ``,
		Parameters: neoism.Props{"fromDateEpoch": fromDateEpoch, "toDateEpoch": toDateEpoch, "mentionsLimit": limit},
		Result:     &results,
	}

	err = pcw.db.Cypher(query)
	if err != nil {
		log.Errorf("Error finding %v most mentioned people between %v and %v with the following statement: %v  Error: %v", limit, fromDateEpoch, toDateEpoch, query.Statement, err)
		return People{}, false, fmt.Errorf("Error finding %v most mentioned people between %v and %v", limit, fromDateEpoch, toDateEpoch)
	}
	log.Debugf("CypherResult MostMentioned was (fromDate=%v, toDate=%v): %+v", limit, fromDateEpoch, toDateEpoch, results)
	if (len(results)) == 0 || len(results[0].Rs) == 0 {
		return People{}, false, nil
	}

	people = neoReadStructToThing(results[0].Rs[0], pcw.env)
	log.Debugf("Returning %v", people)
	return people, true, nil
}

func neoReadStructToThing(neo neoReadStruct, env string) People {
	public := People{}
	return public
}
