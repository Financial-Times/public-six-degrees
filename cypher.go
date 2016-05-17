package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
)

// Driver interface
type Driver interface {
	MostMentioned(fromDate string, toDate string, limit int64) (thing Thing, found bool, err error)
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

func (pcw CypherDriver) MostMentioned(numberOfMostMentioned int64, timePeriod int64) (Thing, bool, error) {
	thing := Thing{}
	results := []struct {
		Rs []neoReadStruct
	}{}
	query := &neoism.CypherQuery{
		Statement:  ``,
		Parameters: neoism.Props{"x": numberOfMostMentioned, "y": timePeriod},
		Result:     &results,
	}

	err := pcw.db.Cypher(query)
	if err != nil {
		log.Errorf("Error finding %v most mentioned people in time period %v with the following statement: %v  Error: %v", numberOfMostMentioned, timePeriod, query.Statement, err)
		return Thing{}, false, fmt.Errorf("Error finding %v most mentioned people in time period %v", numberOfMostMentioned, timePeriod)
	}
	log.Debugf("CypherResult MostMentioned was (x=%v, y=%v): %+v", numberOfMostMentioned, timePeriod, results)
	if (len(results)) == 0 || len(results[0].Rs) == 0 {
		return Thing{}, false, nil
	}

	thing = neoReadStructToThing(results[0].Rs[0], pcw.env)
	log.Debugf("Returning %v", thing)
	return thing, true, nil
}

func neoReadStructToThing(neo neoReadStruct, env string) Thing {
	public := Thing{}
	return public
}
