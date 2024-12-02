package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jacobbrewer1/puppet-reporter/pkg/codegen/apis/api"
	"github.com/jacobbrewer1/uhttp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getRouter(service api.ServerInterface, opts ...api.ServerOption) (*mux.Router, error) {
	r := mux.NewRouter()

	r.HandleFunc("/metrics", uhttp.InternalOnly(promhttp.Handler())).Methods(http.MethodGet)

	r.NotFoundHandler = uhttp.NotFoundHandler()
	r.MethodNotAllowedHandler = uhttp.MethodNotAllowedHandler()

	api.RegisterUnauthedHandlers(
		r,
		service,
		opts...,
	)

	return r, nil
}
