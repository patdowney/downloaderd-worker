package rethinkdb

import (
	r "github.com/dancannon/gorethink"
)

func URLETagIndex(row r.Term) interface{} {
	return []interface{}{
		row.Field("URL"),
		row.Field("Metadata").Field("ETag")}
}
