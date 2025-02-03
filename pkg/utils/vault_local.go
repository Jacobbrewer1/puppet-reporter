//go:build local

package utils

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jacobbrewer1/vaulty"
	"github.com/spf13/viper"
)

func GetVaultClient(ctx context.Context, v *viper.Viper) (vaulty.Client, error) {
	addr := v.GetString("vault.address")
	if addr == "" {
		slog.Info(fmt.Sprintf("No vault address provided, defaulting to %s", defaultVaultAddr))
		addr = defaultVaultAddr
	}

	vc, err := vaulty.NewClient(
		vaulty.WithContext(ctx),
		vaulty.WithAddr(addr),
		vaulty.WithUserPassAuth(v.GetString("vault.username"), v.GetString("vault.password")),
		vaulty.WithKvv2Mount(v.GetString("vault.kvv2_mount")),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating vault client: %w", err)
	}

	return vc, nil
}
