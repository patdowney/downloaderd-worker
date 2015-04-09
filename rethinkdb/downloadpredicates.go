package rethinkdb

import (
	"time"

	r "github.com/dancannon/gorethink"
)

func IsIncomplete() r.Term {
	return r.Row.Field("Metadata").Field("Size").Ne(r.Row.Field("Status").Field("BytesRead"))
}

func NotStarted() r.Term {
	var time time.Time
	return r.Row.Field("TimeStarted").Eq(time)
}

func Started() r.Term {
	var time time.Time
	return r.Row.Field("TimeStarted").Gt(time)
}
