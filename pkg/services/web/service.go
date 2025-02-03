package web

import (
	"embed"
	"net/http"

	"github.com/gorilla/mux"
	repo "github.com/jacobbrewer1/puppet-reporter/pkg/repositories/web"
	"github.com/jacobbrewer1/uhttp"
)

//go:embed templates/*
var localTemplates embed.FS

const (
	htmxHeaderName = "HX-Request"
)

type Service interface {
	Register(r *mux.Router, middleware ...http.HandlerFunc)
}

type service struct {
	r repo.Repository
}

func NewService(r repo.Repository) Service {
	return &service{
		r: r,
	}
}

func (s *service) Register(r *mux.Router, middleware ...http.HandlerFunc) {
	apiRouter := r.PathPrefix("/api").Subrouter()

	r.HandleFunc("/", wrapHandler(s.indexHandler, middleware...)).Methods(http.MethodGet)

	apiRouter.HandleFunc("/reports", wrapHandler(s.APIListReports, middleware...)).Methods(http.MethodGet)
	apiRouter.HandleFunc("/reports/total", wrapHandler(s.APIReportsTotal, middleware...)).Methods(http.MethodGet)
}

func wrapHandler(next http.HandlerFunc, middleware ...http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cw := uhttp.NewResponseWriter(w,
			uhttp.WithDefaultContentType("text/html"),
			uhttp.WithDefaultStatusCode(http.StatusOK),
		)

		for _, m := range middleware {
			m(cw, r)
		}
		next(cw, r)
	}
}
