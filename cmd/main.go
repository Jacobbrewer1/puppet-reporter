package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	main2 "github.com/jacobbrewer1/puppet-reporter/cmd"
	"github.com/jacobbrewer1/puppet-reporter/pkg/logging"
	"github.com/jacobbrewer1/vaulty"
)

var (
	port = flag.Int("port", 8080, "The port to listen on")
)

type App interface {
	Start()
}

type app struct {
	ctx context.Context
	r   *mux.Router
	vc  vaulty.Client
}

func newApp(
	ctx context.Context,
	r *mux.Router,
	vc vaulty.Client,
) App {
	return &app{
		ctx: ctx,
		r:   r,
		vc:  vc,
	}
}

func (a *app) Start() {
	svr := &http.Server{
		Addr:                         fmt.Sprintf(":%d", *port),
		Handler:                      a.r,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  0,
		ReadHeaderTimeout:            0,
		WriteTimeout:                 0,
		IdleTimeout:                  0,
		MaxHeaderBytes:               0,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     nil,
		BaseContext:                  nil,
		ConnContext:                  nil,
	}

	go func() {
		if err := svr.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Error starting server", slog.String(logging.KeyError, err.Error()))
		}
	}()

	<-a.ctx.Done()

	if err := svr.Shutdown(a.ctx); err != nil {
		slog.Error("Error shutting down server", slog.String(logging.KeyError, err.Error()))
		return
	}

	slog.Info("Server shutdown gracefully")
}

func init() {
	flag.Parse()
	main2.initializeLogger()
}

func main() {
}
