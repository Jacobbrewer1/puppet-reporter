package main

import (
	"flag"
	"fmt"

	"github.com/jacobbrewer1/puppet-reporter/pkg/codegen/apis/api"
	"github.com/jacobbrewer1/uhttp"
	"github.com/spf13/viper"
)

const (
	appName = "puppet-reporter"
)

var (
	configLocation = flag.String("config", "config.json", "The location of the config file")
)

func getConfig() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(*configLocation)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	return v, nil
}

func getServerOptions() []api.ServerOption {
	return []api.ServerOption{
		api.WithMetricsMiddleware(metricsMiddleware),
		api.WithErrorHandlerFunc(uhttp.GenericErrorHandler),
	}
}
