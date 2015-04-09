package api

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

// LinkResolver ...
type LinkResolver struct {
	DefaultScheme string
	DefaultHost   string
	Router        *mux.Router
	req           *http.Request
}

func (r *LinkResolver) urlScheme(req *http.Request) string {
	scheme := r.DefaultScheme
	if req != nil {
		scheme = "http"
		if req.TLS != nil {
			scheme = "https"
		}
	}
	return scheme
}

func (r *LinkResolver) urlHost(req *http.Request) string {
	host := r.DefaultHost
	if req != nil {
		host = req.Host
	}
	return host
}

// ResolveLinks ...
func (r *LinkResolver) ResolveLinks(req *http.Request, links *[]Link) {
	for i := range *links {
		r.ResolveLink(req, &(*links)[i])
	}
}

// ResolveLink ...
func (r *LinkResolver) ResolveLink(req *http.Request, link *Link) {
	route := r.Router.Get(link.RouteName)
	if route != nil {
		u, err := route.URL(link.ValueID, link.Value)
		if err != nil {
			log.Print(err)
		} else {
			u.Host = r.urlHost(req)
			u.Scheme = r.urlScheme(req)
			link.Href = u.String()
		}
	}
}

// NewLinkResolver ...
func NewLinkResolver(router *mux.Router) *LinkResolver {
	r := LinkResolver{Router: router, DefaultScheme: "http", DefaultHost: "localhost:8080"}
	return &r
}
