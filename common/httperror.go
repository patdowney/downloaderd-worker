package common

import (
	"fmt"
)

type HTTPError struct {
	URL        string
	Method     string
	StatusCode int
	Status     string
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("%s failed for %s with code %d (%s)", e.Method, e.URL, e.StatusCode, e.Status)
}
