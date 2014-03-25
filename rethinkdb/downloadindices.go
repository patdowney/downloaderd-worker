package rethinkdb

import (
	r "github.com/dancannon/gorethink"
)

func URLETagIndex(row r.RqlTerm) interface{} {
	return []interface{}{
		row.Field("URL"),
		row.Field("Metadata").Field("ETag")}
}
