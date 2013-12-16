# WoW ElasticSearch Tests

A simple example of importing WoW Item Data via the public Item API at http://blizzard.github.io/api-wow-docs/#item-api/individual-item to a Couchbase cluster and then using ElasticSearch to search for items.  Additionally, a node.js import is available to save the source documents directly to ElasticSearch.

NOTE: I changed the default Couchbase elasticsearch transport plugin so the _source on any new indices would also include doc.* (which avoids the "roundtrip" back to Couchbase)

