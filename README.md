# WoW Couchbase Tests

A simple example of importing WoW Item Data via the public Item API at http://blizzard.github.io/api-wow-docs/#item-api/individual-item to a Couchbase cluster and then using ElasticSearch to search for items.  

NOTE: I changed the default Couchbase elasticsearch transport plugin so the _source on any new indices would also include doc.* (which avoids the "roundtrip" back to Couchbase)

# Demo
Check it out at: http://rdp.arcaneorb.com:8093/ (if I have chosen to keep it running)

# Outline 

1) Download all Item Static Data via
http://us.battle.net/api/wow/item/{id}

2) Store all Documents directly in Couchbase

3) Use ElasticSearch to Index everything

4) Run some tests!