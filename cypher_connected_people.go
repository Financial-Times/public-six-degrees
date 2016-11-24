package main

import (
	"fmt"

	"github.com/Financial-Times/neo-model-utils-go/mapper"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
)

type neoContentReadStruct struct {
	UUID      string `json:"uuid"`
	PrefLabel string `json:"prefLabel"`
}

type neoConnectedPeopleReadStruct struct {
	UUID        string                 `json:"uuid"`
	PrefLabel   string                 `json:"prefLabel"`
	Count       int                    `json:"count"`
	ContentList []neoContentReadStruct `json:"contentList"`
}

func (pcw CypherDriver) ConnectedPeople(uuid string, fromDateEpoch int64, toDateEpoch int64, limit int, minimumConnections int, contentLimit int) (connectedPeople []ConnectedPerson, found bool, err error) {
	results := []neoConnectedPeopleReadStruct{}

	statement := `
	MATCH (c:Content) where c.publishedDateEpoch < {toDate} and c.publishedDateEpoch > {fromDate}
	WITH c
	MATCH (p:Person{uuid:{uuid}})<-[:MENTIONS]-(c)-[:MENTIONS]->(p2:Person)
	WITH p, count(distinct(c)) as cm, p2, collect({uuid: c.uuid, prefLabel: c.prefLabel})[0..{contentLimit}] as content
	WHERE cm > {minimumConnections}
	WITH p2.uuid as uuid, p2.prefLabel as prefLabel, cm as count, content as contentList
	RETURN prefLabel, uuid, count, contentList
	ORDER BY count DESC LIMIT {limit}`
	//thing := Thing{}
	query := &neoism.CypherQuery{
		Statement: statement,
		Parameters: neoism.Props{
			"uuid":               uuid,
			"fromDate":           fromDateEpoch,
			"toDate":             toDateEpoch,
			"minimumConnections": minimumConnections,
			"limit":              limit,
			"contentLimit":       contentLimit,
		},
		Result: &results,
	}
	err = pcw.db.Cypher(query)
	if err != nil {
		log.Errorf(`Error finding people with more than %v connections to person with uuid %v
      between %v and %v with the following statement: %v  Error: %v`, limit, uuid, fromDateEpoch, toDateEpoch, query.Statement, err)
		return []ConnectedPerson{}, false, fmt.Errorf("Error finding people with more than %v connections to person with uuid %v between %v and %v with the following statement: %v  Error: %v", limit, uuid, fromDateEpoch, toDateEpoch, query.Statement, err)
	}

	if (len(results)) == 0 {
		return []ConnectedPerson{}, false, nil
	}

	connectedPeopleResults := neoReadStructToConnectedPeople(&results, pcw.env)

	return connectedPeopleResults, true, nil
}

func neoReadStructToConnectedPeople(neo *[]neoConnectedPeopleReadStruct, env string) []ConnectedPerson {
	connectedPeople := []ConnectedPerson{}
	for _, neoCP := range *neo {
		var connectedPerson = ConnectedPerson{}
		connectedPerson.Person.APIURL = mapper.APIURL(neoCP.UUID, []string{"Person"}, env)
		connectedPerson.Person.ID = mapper.IDURL(neoCP.UUID)
		connectedPerson.Person.PrefLabel = neoCP.PrefLabel
		connectedPerson.Count = neoCP.Count

		contentList := []Content{}

		for _, neoContent := range neoCP.ContentList {
			var content = Content{}
			content.ID = neoContent.UUID
			content.Title = neoContent.PrefLabel
			content.APIURL = mapper.APIURL(neoContent.UUID, []string{"Content"}, env)
			contentList = append(contentList, content)
		}
		connectedPerson.Content = contentList
		connectedPeople = append(connectedPeople, connectedPerson)
	}
	return connectedPeople
}
