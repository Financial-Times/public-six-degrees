package sixdegrees

import (
	"github.com/Financial-Times/neo-model-utils-go/mapper"
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

func (cd CypherDriver) ConnectedPeople(uuid string, fromDateEpoch int64, toDateEpoch int64, resultLimit int, minimumConnections int, contentLimit int) ([]ConnectedPerson, bool, error) {
	results := []neoConnectedPeopleReadStruct{}

	statement := `
		MATCH (c:Content)
		WHERE
			c.publishedDateEpoch < {toDate}
			AND c.publishedDateEpoch > {fromDate}
		MATCH (p:Person{prefUUID:{uuid}})<-[:EQUIVALENT_TO]-(:Person)<-[:MENTIONS]-(c)
		MATCH (c)-[:MENTIONS]->(:Person)-[:EQUIVALENT_TO]->(p2:Person)
		WITH
			c,
			p,
			p2
		ORDER BY
			c.uuid DESC
		WITH
			p,
			count(distinct(c)) as cm,
			p2,
			collect({
				uuid: c.uuid,
				prefLabel: c.prefLabel
			})[0..{contentLimit}] as content
		WHERE cm >= {minimumConnections}
		WITH
			p2.prefUUID as uuid,
			p2.prefLabel as prefLabel,
			cm as count,
			content as contentList
		RETURN
			prefLabel,
			uuid,
			count,
			contentList
		ORDER BY
			count DESC,
			uuid ASC
		LIMIT {limit}
	`

	query := &neoism.CypherQuery{
		Statement: statement,
		Parameters: neoism.Props{
			"uuid":               uuid,
			"fromDate":           fromDateEpoch,
			"toDate":             toDateEpoch,
			"minimumConnections": minimumConnections,
			"limit":              resultLimit,
			"contentLimit":       contentLimit,
		},
		Result: &results,
	}

	if err := cd.conn.CypherBatch([]*neoism.CypherQuery{query}); err != nil || len(results) == 0 {
		return []ConnectedPerson{}, false, err
	}

	return transformToConnectedPeople(&results), true, nil
}

func transformToConnectedPeople(neo *[]neoConnectedPeopleReadStruct) []ConnectedPerson {
	connectedPeople := []ConnectedPerson{}
	for _, neoCP := range *neo {
		var connectedPerson = ConnectedPerson{}
		connectedPerson.Person.APIURL = mapper.APIURL(neoCP.UUID, []string{"Person"}, "local")
		connectedPerson.Person.ID = mapper.IDURL(neoCP.UUID)
		connectedPerson.Person.PrefLabel = neoCP.PrefLabel
		connectedPerson.Count = neoCP.Count

		contentList := []Content{}

		for _, neoContent := range neoCP.ContentList {
			var content = Content{}
			content.ID = neoContent.UUID
			content.Title = neoContent.PrefLabel
			content.APIURL = mapper.APIURL(neoContent.UUID, []string{"Content"}, "local")
			contentList = append(contentList, content)
		}
		connectedPerson.Content = contentList
		connectedPeople = append(connectedPeople, connectedPerson)
	}
	return connectedPeople
}
