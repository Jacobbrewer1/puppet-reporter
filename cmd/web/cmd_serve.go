package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"

	"github.com/google/subcommands"
	"github.com/gorilla/mux"
	"github.com/jacobbrewer1/puppet-reporter/pkg/logging"
	repo "github.com/jacobbrewer1/puppet-reporter/pkg/repositories/web"
	"github.com/jacobbrewer1/puppet-reporter/pkg/services/web"
	"github.com/jacobbrewer1/puppet-reporter/pkg/utils"
	"github.com/jacobbrewer1/uhttp"
	"github.com/jacobbrewer1/vaulty"
	"github.com/jacobbrewer1/vaulty/repositories"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

type serveCmd struct {
	// port is the port to listen on
	port string

	// configLocation is the location of the config file
	configLocation string
}

func (s *serveCmd) Name() string {
	return "serve"
}

func (s *serveCmd) Synopsis() string {
	return "Start the web server"
}

func (s *serveCmd) Usage() string {
	return `serve:
  Start the web server.
`
}

func (s *serveCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&s.port, "port", "8080", "The port to listen on")
	f.StringVar(&s.configLocation, "config", "config.json", "The location of the config file")
}

func (s *serveCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	r := mux.NewRouter()
	err := s.setup(ctx, r)
	if err != nil {
		slog.Error("Error setting up server", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	slog.Info(
		"Starting application",
		slog.String("version", Commit),
		slog.String("runtime", fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)),
		slog.String("build_date", Date),
	)

	srv := &http.Server{
		Addr:    ":" + s.port,
		Handler: r,
	}

	// Start the server in a goroutine, so we can listen for the context to be done.
	go func(srv *http.Server) {
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			slog.Info("Server closed gracefully")
			os.Exit(0)
		} else if err != nil {
			slog.Error("Error serving requests", slog.String(logging.KeyError, err.Error()))
			os.Exit(1)
		}
	}(srv)

	<-ctx.Done()
	slog.Info("Shutting down application")
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Error shutting down application", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func (s *serveCmd) setup(ctx context.Context, r *mux.Router) (err error) {
	v := viper.New()
	v.SetConfigFile(s.configLocation)
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	if !v.IsSet("vault") {
		return errors.New("vault configuration not found")
	}

	slog.Info("Vault configuration found, attempting to connect")

	vc, err := utils.GetVaultClient(ctx, v)
	if err != nil {
		return fmt.Errorf("error creating vault client: %w", err)
	}

	slog.Debug("Vault client created")

	vs, err := vc.Path(v.GetString("vault.database.role"), vaulty.WithPrefix(v.GetString("vault.database.path"))).GetSecret(ctx)
	if errors.Is(err, vaulty.ErrSecretNotFound) {
		return fmt.Errorf("secrets not found in vault: %s", v.GetString("vault.database.path"))
	} else if err != nil {
		return fmt.Errorf("error getting secrets from vault: %w", err)
	}

	dbConnector, err := repositories.NewDatabaseConnector(
		repositories.WithContext(ctx),
		repositories.WithVaultClient(vc),
		repositories.WithCurrentSecrets(vs),
		repositories.WithViper(v),
	)
	if err != nil {
		return fmt.Errorf("error creating database connector: %w", err)
	}

	db, err := dbConnector.ConnectDB()
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	slog.Info("Database connection generate from vault secrets")

	repository := repo.NewRepository(db)
	web.NewService(repository).Register(r)

	r.HandleFunc("/metrics", uhttp.InternalOnly(promhttp.Handler())).Methods(http.MethodGet)
	r.HandleFunc("/health", uhttp.InternalOnly(healthHandler(db))).Methods(http.MethodGet)

	r.NotFoundHandler = uhttp.NotFoundHandler()
	r.MethodNotAllowedHandler = uhttp.MethodNotAllowedHandler()

	return nil
}
