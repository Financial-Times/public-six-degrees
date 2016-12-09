package main

import (
	"github.com/Financial-Times/neo-utils-go/neoutils"
)

type driver interface {
	ConnectedPeople(uuid string, fromDateEpoch int64, toDateEpoch int64, limit int, minimumConnections int, contentLimit int) (connectedPeople []ConnectedPerson, found bool, err error)
	MostMentioned(fromDateEpoch int64, toDateEpoch int64, limit int) (thingList []Thing, found bool, err error)
	CheckConnectivity() error
}

type cypherDriver struct {
	conn neoutils.NeoConnection
	env  string
}

//NewCypherDriver instantiate driver
func NewCypherDriver(conn neoutils.NeoConnection, env string) cypherDriver {
	return cypherDriver{conn, env}
}

// CheckConnectivity tests neo4j by running a simple cypher query
func (cd cypherDriver) CheckConnectivity() error {
	return neoutils.Check(cd.conn)
}

type neoMentionsReadStruct struct {
	UUID      string `json:"uuid"`
	PrefLabel string `json:"prefLabel"`
	Mentions  int    `json:"mentions"`
}
