package api

type Link struct {
	Relation  string  `json:"rel"`
	Href      string  `json:"href"`
	MediaType *string `json:"type,omitempty"`
}
