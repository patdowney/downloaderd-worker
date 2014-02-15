package api

type Error struct {
	Error string `json:"error"`
}

func NewError(e error) *Error {
	err := Error{Error: e.Error()}

	return &err
}
