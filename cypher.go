package main

import (
	"github.com/Financial-Times/neo-utils-go/neoutils"
)

type driver interface {
	ConnectedPeople(uuid string, fromDateEpoch int64, toDateEpoch int64, limit int, minimumConnections int, contentLimit int) ([]ConnectedPerson, bool, error)
	MostMentioned(fromDateEpoch int64, toDateEpoch int64, limit int) ([]Thing, bool, error)
	CheckConnectivity() error
}

type cypherDriver struct {
	conn neoutils.NeoConnection
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
