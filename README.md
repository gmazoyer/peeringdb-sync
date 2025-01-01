# PeeringDB Synchronization

This tool is used to synchronize the PeeringDB database in a local SQLite3
database. It can be useful when you need to query the PeeringDB API a lot.
Querying the local database instead will make things much faster.

It also serves as an example of the
[PeeringDB API Go package](https://godoc.org/github.com/gmazoyer/peeringdb).
