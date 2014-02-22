package api

type Link struct {
	Relation  string `json:"rel"`
	Href      string `json:"href"`
	Value     string `json:"-"`
	ValueId   string `json:"-"`
	RouteName string `json:"-"`
}
