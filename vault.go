package main

import (
	"fmt"
	"log/slog"

	"github.com/jacobbrewer1/vaulty"
	"github.com/spf13/viper"
)

func getVaultClient(v *viper.Viper) (vaulty.Client, error) {
	addr := v.GetString("vault.address")
	if addr == "" {
		slog.Info(fmt.Sprintf("No vault address provided, defaulting to %s", defaultVaultAddr))
		addr = defaultVaultAddr
	}

	vc, err := vaulty.NewClient(
		vaulty.WithAddr(addr),
		vaulty.WithKubernetesAuthDefault(),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating vault client: %w", err)
	}

	return vc, nil
}
