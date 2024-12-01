//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/jacobbrewer1/puppet-reporter/pkg/services/api"
)

func InitializeApp() (App, error) {
	wire.Build(
		getRootContext,
		getConfig,
		getVaultClient,
		api.NewService,
		getServerOptions,
		getRouter,
		newApp,
	)
	return new(app), nil
}
