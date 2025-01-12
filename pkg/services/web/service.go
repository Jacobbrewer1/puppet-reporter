package web

import (
	"embed"
	"net/http"

	"github.com/gorilla/mux"
	repo "github.com/jacobbrewer1/puppet-reporter/pkg/repositories/web"
)

//go:embed templates/*.gohtml
var templates embed.FS

type Service interface {
	Register(r *mux.Router, middlewares ...http.HandlerFunc)
}

type service struct {
	// r is the repository used by the service.
	r repo.Repository
}

func NewService(r repo.Repository) Service {
	return &service{
		r: r,
	}
}

func (s *service) Register(r *mux.Router, middlewares ...http.HandlerFunc) {
	r.HandleFunc("/", wrapHandler(s.indexHandler, middlewares...)).Methods(http.MethodGet)
}

func wrapHandler(h http.HandlerFunc, middlewares ...http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, m := range middlewares {
			m(w, r)
		}
		h(w, r)
	}
}
