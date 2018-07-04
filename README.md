[![CircleCI](https://circleci.com/gh/Financial-Times/public-six-degrees/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/public-six-degrees/tree/master)
[![Coverage Status](https://coveralls.io/repos/github/Financial-Times/public-six-degrees/badge.svg)](https://coveralls.io/github/Financial-Times/public-six-degrees)
# Public Six Degrees API

Provides a public API for retrieving most mentioned people and connected people for the 'Six degrees' application.
Connects to Neo4j to get the needed data.

## Installation & running locally

* `go get -u -t github.com/Financial-Times/public-six-degrees`
* `cd $GOPATH/src/github.com/Financial-Times/public-six-degrees`
* `go test -race -v ./...`
*  `./public-six-degrees --neo-url={neo4jUrl}`

## Endpoints
### GET

* `/sixdegrees/connectedPeople` - Get connected people to a given person
    * `uuid` - (required) The given person's UUID we want to query
    * `fromDate` - Start date, in YYYY-MM-DD format. Defaults to one week ago if not given. 
    If toDate is before fromDate, fromDate changes to be a week from toDate. 
    If the difference between fromDate and toDate is greater than 1 year, toDate is changed to be 1 year after fromDate
    * `toDate` - End date, in YYYY-MM-DD format. Defaults to today if not given
    If toDate is before fromDate, fromDate changes to be a week from toDate. 
    If the difference between fromDate and toDate is greater than 1 year, toDate is changed to be 1 year after fromDate     
    * `minimumConnections` - The minimum number of connections required for a connection to appear in the list. Defaults to 5 if not given
    * `contentLimit` - The maximum number of content returned for a mentioned connected person. Defaults to 3 if not given
    * `limit` - The maximum number of resulting connected people. Defaults to 10 if not given
* `/sixdegrees/mostMentionedPeople`
    * `fromDate` - Start date, in YYYY-MM-DD format. Defaults to one week ago if not given
    If toDate is before fromDate, fromDate changes to be a week from toDate. 
    If the difference between fromDate and toDate is greater than 1 year, toDate is changed to be 1 year after fromDate     
    * `toDate` - End date, in YYYY-MM-DD format. Defaults to today if not given 
    If toDate is before fromDate, fromDate changes to be a week from toDate. 
    If the difference between fromDate and toDate is greater than 1 year, toDate is changed to be 1 year after fromDate    
    * `limit` - The maximum number of resulting most mentioned people. Defaults to 20 if not given

### Admin
    
* `/__health`
* `/__gtg`
* `/__ping`
* `/__build-info`
* `/ping`
* `/build-info`

    
## Example

* With `/sixdegrees/connectedPeople`
`GET /sixdegrees/connectedPeople?uuid=dc278df2-1c8b-3e44-8ca8-5d255f75f737&fromDate=2016-01-01&toDate=2016-05-17&limit=2`
```
[{
    "person": {
        "id": "http://api.ft.com/things/9185a2a9-1545-302b-9a16-c63986b67be3",
        "apiUrl": "http://api.ft.com/people/9185a2a9-1545-302b-9a16-c63986b67be3",
        "prefLabel": "Boris Johnson"
    },
    "count": 162,
    "content": [{
        "id": "40b38230-c101-11e5-9fdb-87b8d15baec2",
        "apiUrl": "http://api.ft.com/content/40b38230-c101-11e5-9fdb-87b8d15baec2",
        "title": "Heathrow decision put off until after EU referendum"
    }, {
        "id": "6db05608-18e7-11e6-b197-a4af20d5575e",
        "apiUrl": "http://api.ft.com/content/6db05608-18e7-11e6-b197-a4af20d5575e",
        "title": "Do we still need to build a third runway at Heathrow?"
    }, {
        "id": "81665806-e23b-11e5-9217-6ae3733a2cd1",
        "apiUrl": "http://api.ft.com/content/81665806-e23b-11e5-9217-6ae3733a2cd1",
        "title": "Osborne plays it safe over pensions"
    }]
}, {
    "person": {
        "id": "http://api.ft.com/things/9421d9ee-7e0f-3f7c-8adc-ded83fabdb92",
        "apiUrl": "http://api.ft.com/people/9421d9ee-7e0f-3f7c-8adc-ded83fabdb92",
        "prefLabel": "George Gideon Oliver Osborne"
    },
    "count": 136,
    "content": [{
        "id": "40b38230-c101-11e5-9fdb-87b8d15baec2",
        "apiUrl": "http://api.ft.com/content/40b38230-c101-11e5-9fdb-87b8d15baec2",
        "title": "Heathrow decision put off until after EU referendum"
    }, {
        "id": "6db05608-18e7-11e6-b197-a4af20d5575e",
        "apiUrl": "http://api.ft.com/content/6db05608-18e7-11e6-b197-a4af20d5575e",
        "title": "Do we still need to build a third runway at Heathrow?"
    }, {
        "id": "ceeaecdc-e12c-11e5-9217-6ae3733a2cd1",
        "apiUrl": "http://api.ft.com/content/ceeaecdc-e12c-11e5-9217-6ae3733a2cd1",
        "title": "Minister takes on Treasury over pension Isa"
    }]
}]
```

* With `/sixdegrees/mostMentionedPeople`

`GET /sixdegrees/mostMentionedPeople?fromDate=2016-01-01&toDate=2016-01-02&limit=5`
```
[{
    "id": "http://api.ft.com/things/dc278df2-1c8b-3e44-8ca8-5d255f75f737",
    "prefLabel": "David William Donald Cameron"
}, {
    "id": "http://api.ft.com/things/8d9470c9-127e-3fc7-95a0-71804cc5ea9d",
    "prefLabel": "Hillary Rodham Clinton"
}, {
    "id": "http://api.ft.com/things/3ead3886-85d3-36ee-95e3-75ed1dc832b7",
    "prefLabel": "John Ellis Bush"
}, {
    "id": "http://api.ft.com/things/4b600f39-7706-3acd-897d-3b81100b30bd",
    "prefLabel": "Joachim Herrmann"
}, {
    "id": "http://api.ft.com/things/889a4f48-c4df-3e2e-89a5-7665b49ced07",
    "prefLabel": "Jack A. Ablin"
}]
```
