package rethinkdb

import (
	"time"

	r "github.com/dancannon/gorethink"
)

func IsNotFinished() r.RqlTerm {
	return r.Row.Field("Finished").Eq(false)
}

func IsFinished() r.RqlTerm {
	return r.Row.Field("Finished").Eq(true)
}

func IsWaiting() r.RqlTerm {
	var time time.Time
	return r.Row.Field("TimeStarted").Lt(time).And(IsNotFinished())
}

func InProgress() r.RqlTerm {
	var time time.Time
	return r.Row.Field("TimeStarted").Gt(time).And(IsNotFinished())
}
