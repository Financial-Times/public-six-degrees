package sixdegrees

import (
	"github.com/Financial-Times/neo-model-utils-go/mapper"
	"github.com/jmcvetta/neoism"
)

func (cd CypherDriver) MostMentioned(fromDateEpoch int64, toDateEpoch int64, limit int) ([]Thing, bool, error) {
	results := []neoMentionsReadStruct{}
	query := &neoism.CypherQuery{
		Statement: `MATCH (c:Content)-[a:MENTIONS]->(:Person)-[:EQUIVALENT_TO]->(p:Person)
					WHERE c.publishedDateEpoch > {fromDateEpoch} AND c.publishedDateEpoch < {toDateEpoch}
					WITH p.prefLabel as prefLabel, p.prefUUID as uuid,
					COUNT(a) as mentions
					RETURN uuid, prefLabel, mentions
					ORDER BY mentions
					DESC LIMIT {mentionsLimit}`,
		Parameters: neoism.Props{"fromDateEpoch": fromDateEpoch, "toDateEpoch": toDateEpoch, "mentionsLimit": limit},
		Result:     &results,
	}

	if err := cd.conn.CypherBatch([]*neoism.CypherQuery{query}); err != nil || len(results) == 0 {
		return []Thing{}, false, err
	}

	return transformToMentionPeople(&results), true, nil
}

func transformToMentionPeople(neo *[]neoMentionsReadStruct) []Thing {
	peopleList := []Thing{}
	for _, neoCon := range *neo {
		var thing = Thing{}
		thing.ID = mapper.IDURL(neoCon.UUID)
		thing.PrefLabel = neoCon.PrefLabel
		peopleList = append(peopleList, thing)
	}
	return peopleList
}
