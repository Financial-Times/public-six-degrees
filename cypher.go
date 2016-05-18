package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
)

// Driver interface
type Driver interface {
	ConnectedPeople(uuid string, fromDateEpoch int64, toDateEpoch int64, limit int, minimumConnections int) (connectedPeople []ConnectedPerson, found bool, err error)
	MostMentioned(fromDateEpoch int64, toDateEpoch int64, limit int) (thingList []Thing, found bool, err error)
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

type neoMentionsReadStruct struct {
	UUID      string `json:"uuid"`
	PrefLabel string `json:"prefLabel"`
	Mentions  int    `json:"mentions"`
}
