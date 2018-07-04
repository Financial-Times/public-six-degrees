package sixdegrees

import (
	"github.com/Financial-Times/neo-utils-go/neoutils"
)

type Driver interface {
	ConnectedPeople(uuid string, fromDateEpoch int64, toDateEpoch int64, limit int, minimumConnections int, contentLimit int) ([]ConnectedPerson, bool, error)
	MostMentioned(fromDateEpoch int64, toDateEpoch int64, limit int) ([]Thing, bool, error)
	CheckConnectivity() error
}

func NewCypherDriver(conn neoutils.NeoConnection) Driver {
	return &CypherDriver{
		conn: conn,
	}
}

type CypherDriver struct {
	conn neoutils.NeoConnection
}

func (cd CypherDriver) CheckConnectivity() error {
	return neoutils.Check(cd.conn)
}

type neoMentionsReadStruct struct {
	UUID      string `json:"uuid"`
	PrefLabel string `json:"prefLabel"`
	Mentions  int    `json:"mentions"`
}
