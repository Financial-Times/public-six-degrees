package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
)

// Driver interface
type Driver interface {
	MostMentioned(fromDate string, toDate string, limit int64) (thing Thing, found bool, err error)
	ConnectedPeople(uuid string, fromDate string, toDate string, limit int64, minimumConnections int64) (connectedPeople[] ConnectedPerson, found bool, err error)
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


func (pcw CypherDriver) MostMentioned(fromDate string, toDate string, limit int64) (Thing, bool, error) {
	return Thing{"bla", "bla", "bla"}, true, nil
}

//MATCH (c:Content) where c.publishedDateEpoch < 1460371050 and c.publishedDateEpoch > 1457605324
//WITH c
//MATCH (p:Person{uuid:"9421d9ee-7e0f-3f7c-8adc-ded83fabdb92"})<-[:MENTIONS]-(c)-[:MENTIONS]->(p2:Person)
//WITH p, count(distinct(c)) as cm, p2
//WHERE cm > 5
//return cm, p2 limit 25
func (pcw CypherDriver) ConnectedPeople(uuid string, fromDate string, toDate string, limit int64, minimumConnections int64) (Thing, bool, error) {

	statement := `
	MATCH (c:Content) where c.publishedDateEpoch < {toDate} and c.publishedDateEpoch > {fromDate}
	WITH c
	MATCH (p:Person{uuid:{uuid}})<-[:MENTIONS]-(c)-[:MENTIONS]->(p2:Person)
	WITH p, count(distinct(c)) as cm, p2
	WHERE cm > {minimumConnections}
	WITH p2.uuid as uuid
	RETURN name, uuid, mentions ORDER BY mentions DESC LIMIT {limit}`
	thing := Thing{}
	results := []struct {
		Rs []neoReadStruct
	}{}
	query := &neoism.CypherQuery{
		Statement:  statement,
		Parameters: neoism.Props{
			"uuid": uuid,
			"fromDate": fromDate,
			"toDate": toDate,
			"minimumConnections": minimumConnections,
			"limit": limit,
		},
		Result:     &results,
	}
	err := pcw.db.Cypher(query)
	if err != nil {
		log.Errorf("Error finding %v most mentioned people in time period %v->%v with the following statement: %v  Error: %v", limit, fromDate, toDate, query.Statement, err)
		return Thing{}, false, fmt.Errorf("Error finding %v most mentioned people in time period %v to %v", limit, fromDate, toDate)
	}
	log.Debugf("CypherResult MostMentioned was (limit=%v, fromDate=%v, toDate=%v): %+v", limit, fromDate, toDate, results)
	if (len(results)) == 0 || len(results[0].Rs) == 0 {
		return Thing{}, false, nil
	}

	thing = neoReadStructToThing(results[0].Rs[0], pcw.env)
	log.Debugf("Returning %v", thing)
	return thing, true, nil
}

// MATCH (c:Content)-[a:MENTIONS]->(p:Person)
// WHERE c.publishedDateEpoch > {publishedDateEpoch}
// WITH p.prefLabel as name, p.uuid as uuid, COUNT(a) as mentions
// RETURN name, uuid, mentions ORDER BY mentions DESC LIMIT 20
//func (pcw CypherDriver) MostMentioned(fromDate string, toDate string, limit int64) (Thing, bool, error) {
//	thing := Thing{}
//	results := []struct {
//		Rs []neoReadStruct
//	}{}
//	query := &neoism.CypherQuery{
//		Statement:  ``,
//		//Parameters: neoism.Props{"x": limit, "y": timePeriod},
//		Parameters: neoism.Props{"x": limit}
//		Result:     &results,
//	}
//
//	err := pcw.db.Cypher(query)
//	if err != nil {
//		log.Errorf("Error finding %v most mentioned people in time period %v with the following statement: %v  Error: %v", limit, timePeriod, query.Statement, err)
//		return Thing{}, false, fmt.Errorf("Error finding %v most mentioned people in time period %v", limit, timePeriod)
//	}
//	log.Debugf("CypherResult MostMentioned was (x=%v, y=%v): %+v", limit, timePeriod, results)
//	if (len(results)) == 0 || len(results[0].Rs) == 0 {
//		return Thing{}, false, nil
//	}
//
//	thing = neoReadStructToThing(results[0].Rs[0], pcw.env)
//	log.Debugf("Returning %v", thing)
//	return thing, true, nil
//}

func neoReadStructToThing(neo neoReadStruct, env string) Thing {
	public := Thing{}
	return public
}
