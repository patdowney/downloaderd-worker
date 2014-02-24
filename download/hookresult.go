package download

import (
	"time"
)

type HookResult struct {
	Errors     []string
	StatusCode int
	Time       time.Time
}

func (r *HookResult) AddError(err error) {
	r.Errors = append(r.Errors, err.Error())
}

func NewHookResult() *HookResult {
	hr := HookResult{
		Errors: make([]string, 0)}

	return &hr
}
