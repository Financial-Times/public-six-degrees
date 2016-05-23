Public Six Degrees
==================

An implementation of the FT Labs "Six degrees of Angela Merkel" demo that was produced
in 2015. Connects to data in neo4j, and is written in Go.


Getting Started
---------------

1. Download and install neo4j.
1. Disable authentication.
1. Start up neo4j.
1. Build the app:

        go build

2. Run the app (Mac / Linux):

        ./public-six-degrees

3. Visit: [http://localhost:8080/sixdegrees/connectedPeople](http://localhost:8080/sixdegrees/connectedPeople)


Sample queries
--------------

    curl http://sixdegrees-demo.in.ft.com/sixdegrees/connectedPeople?uuid=36c6124-24c0-39fe-9172-d37c60eafdeg&fromDate=2016-05-17&toDate=2016-05-18
    curl http://sixdegrees-demo.in.ft.com/sixdegrees/connectedPeople?uuid=dc278df2-1c8b-3e44-8ca8-5d255f75f737&fromDate=2014-01-01&toDate=2016-05-17&minimumConnections=1
    curl http://sixdegrees-demo.in.ft.com/sixdegrees/mostMentioned?fromDate=2014-01-01&toDate=2016-05-17
    curl http://sixdegrees-demo.in.ft.com/sixdegrees/mostMentioned?fromDate=2014-01-01&toDate=2016-05-17&limit=50
    
    MostMentioned defaults to the last week and limit of 20
    curl http://sixdegrees-demo.in.ft.com/sixdegrees/mostMentioned

API
---

See [swagger.yaml](apidoc/swagger.yaml).


Test Environment
----------------

We have a test environment running in EC2.


References
----------

1. http://ftlabs.github.io/six-degrees/ - original demo
    1. http://ftlabs.github.io/six-degrees/graph.html - the bobbly graph
    1. http://ftlabs.github.io/six-degrees/erdos.html - Merkel chains
    1. http://ftlabs-sapi-capi-slurp-slice.herokuapp.com/display_options
1. http://bl.ocks.org/mbostock/4062045 - d3.js force-directed graph
1. http://editor.swagger.io/#/ - Swagger editor for producing our API docs.
    1. https://github.com/swagger-api/swagger-ui - for displaying API docs
