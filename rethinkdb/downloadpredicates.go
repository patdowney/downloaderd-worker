package rethinkdb

import (
	"time"

	r "github.com/dancannon/gorethink"
)

func IsIncomplete() r.RqlTerm {
	return r.Row.Field("Metadata").Field("Size").Ne(r.Row.Field("Status").Field("BytesRead"))
}

func IsNotFinished() r.RqlTerm {
	return r.Row.Field("Finished").Eq(false)
}

func IsFinished() r.RqlTerm {
	return r.Row.Field("Finished").Eq(true)
}

func IsWaiting() r.RqlTerm {
	var time time.Time
	return r.Row.Field("TimeStarted").Eq(time).And(IsNotFinished())
}

func InProgress() r.RqlTerm {
	var time time.Time
	return r.Row.Field("TimeStarted").Gt(time).And(IsNotFinished())
}
