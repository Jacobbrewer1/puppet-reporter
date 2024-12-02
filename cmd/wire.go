//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/jacobbrewer1/puppet-reporter/cmd"
	"github.com/jacobbrewer1/puppet-reporter/pkg/services/api"
)

func InitializeApp() (main.App, error) {
	wire.Build(
		getRootContext,
		getConfig,
		getVaultClient,
		api.NewService,
		getServerOptions,
		getRouter,
		main.newApp,
	)
	return new(main.app), nil
}
