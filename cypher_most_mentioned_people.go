package main

import (
	"github.com/Financial-Times/neo-model-utils-go/mapper"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
)

func (pcw cypherDriver) MostMentioned(fromDateEpoch int64, toDateEpoch int64, limit int) (thingList []Thing, found bool, err error) {
	log.Infof("logging fromDate:%v toDate:%v  limit:%v", fromDateEpoch, toDateEpoch, limit)
	results := []neoMentionsReadStruct{}
	query := &neoism.CypherQuery{
		Statement: `MATCH (c:Content)-[a:MENTIONS]->(p:Person)
					WHERE c.publishedDateEpoch > {fromDateEpoch} AND c.publishedDateEpoch < {toDateEpoch}
					WITH p.prefLabel as prefLabel, p.uuid as uuid,
					COUNT(a) as mentions
					RETURN uuid, prefLabel, mentions
					ORDER BY mentions
					DESC LIMIT {mentionsLimit}`,
		Parameters: neoism.Props{"fromDateEpoch": fromDateEpoch, "toDateEpoch": toDateEpoch, "mentionsLimit": limit},
		Result:     &results,
	}
	log.Debugf("Query %v", query)

	if err = pcw.conn.CypherBatch([]*neoism.CypherQuery{query}); err != nil {
		log.Errorf("Error finding %v most mentioned people between %v and %v with the following statement: %v  Error: %v", limit, fromDateEpoch, toDateEpoch, query.Statement, err)
		return []Thing{}, false, err
	}
	log.Debugf("CypherResult MostMentioned was (fromDate=%v, toDate=%v)", limit, fromDateEpoch, toDateEpoch)

	thingList, _ = neoReadStructToMentionPeople(&results)
	log.Debugf("Result: %v\n", thingList)
	return thingList, true, nil
}

func neoReadStructToMentionPeople(neo *[]neoMentionsReadStruct) ([]Thing, error) {
	peopleList := []Thing{}
	for _, neoCon := range *neo {
		log.Infof("neoCon result: %v", neoCon)
		var thing = Thing{}
		thing.ID = mapper.IDURL(neoCon.UUID)
		thing.PrefLabel = neoCon.PrefLabel
		peopleList = append(peopleList, thing)
	}
	return peopleList, nil
}
