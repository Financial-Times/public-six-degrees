package main

import (
	"fmt"

	"github.com/Financial-Times/neo-model-utils-go/mapper"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
)

type neoConnectedPeopleReadStruct struct {
	UUID      string `json:"uuid"`
	PrefLabel string `json:"prefLabel"`
	Count     int    `json:"count"`
}

func (pcw CypherDriver) ConnectedPeople(uuid string, fromDateEpoch int64, toDateEpoch int64, limit int, minimumConnections int) (connectedPeople []ConnectedPerson, found bool, err error) {
	results := []neoConnectedPeopleReadStruct{}

	statement := `
	MATCH (c:Content) where c.publishedDateEpoch < {toDate} and c.publishedDateEpoch > {fromDate}
	WITH c
	MATCH (p:Person{uuid:{uuid}})<-[:MENTIONS]-(c)-[:MENTIONS]->(p2:Person)
	WITH p, count(distinct(c)) as cm, p2
	WHERE cm > {minimumConnections}
	WITH p2.uuid as uuid, p2.prefLabel as prefLabel, cm as count
	RETURN prefLabel, uuid, count ORDER BY count DESC LIMIT {limit}`
	//thing := Thing{}
	query := &neoism.CypherQuery{
		Statement: statement,
		Parameters: neoism.Props{
			"uuid":               uuid,
			"fromDate":           fromDateEpoch,
			"toDate":             toDateEpoch,
			"minimumConnections": minimumConnections,
			"limit":              limit,
		},
		Result: &results,
	}
	err = pcw.db.Cypher(query)
	if err != nil {
		log.Errorf(`Error finding people with more than %v connections to person with uuid %v
      between %v and %v with the following statement: %v  Error: %v`, limit, uuid, fromDateEpoch, toDateEpoch, query.Statement, err)
		return []ConnectedPerson{}, false, fmt.Errorf("Error finding people with more than %v connections to person with uuid %v between %v and %v with the following statement: %v  Error: %v", limit, uuid, fromDateEpoch, toDateEpoch, query.Statement, err)
	}
	log.Infof("CypherResult connectedPeople was (fromDate=%v, toDate=%v): %+v", fromDateEpoch, toDateEpoch, results)
	if (len(results)) == 0 {
		return []ConnectedPerson{}, false, nil
	}

	connectedPeopleResults := neoReadStructToConnectedPeople(&results, pcw.env)

	log.Infof("Returning %v", connectedPeopleResults)
	return connectedPeopleResults, true, nil
}

func neoReadStructToConnectedPeople(neo *[]neoConnectedPeopleReadStruct, env string) []ConnectedPerson {
	connectedPeople := []ConnectedPerson{}
	for _, neoCP := range *neo {
		var connectedPerson = ConnectedPerson{}
		connectedPerson.Person.APIURL = mapper.APIURL(neoCP.UUID, []string{"person"}, env)
		connectedPerson.Person.ID = mapper.IDURL(neoCP.UUID)
		connectedPerson.Person.PrefLabel = neoCP.PrefLabel
		connectedPerson.Count = neoCP.Count
		connectedPeople = append(connectedPeople, connectedPerson)
	}
	return connectedPeople
}
