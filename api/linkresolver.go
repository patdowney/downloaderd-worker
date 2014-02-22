package api

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type LinkResolver struct {
	Router *mux.Router
	req    *http.Request
}

func (r *LinkResolver) urlScheme(req *http.Request) string {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	return scheme
}

func (r *LinkResolver) ResolveLinks(req *http.Request, links *[]Link) {
	for i, _ := range *links {
		r.ResolveLink(req, &(*links)[i])
	}
}

func (r *LinkResolver) ResolveLink(req *http.Request, link *Link) {
	route := r.Router.Get(link.RouteName)
	if route != nil {
		u, err := route.URL(link.ValueId, link.Value)
		if err != nil {
			log.Print(err)
		} else {
			u.Host = req.Host
			u.Scheme = r.urlScheme(req)
			link.Href = u.String()
		}
	}
}

func NewLinkResolver(router *mux.Router) *LinkResolver {
	r := LinkResolver{Router: router}
	return &r
}
