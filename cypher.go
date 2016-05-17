package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
)

// Driver interface
type Driver interface {
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
