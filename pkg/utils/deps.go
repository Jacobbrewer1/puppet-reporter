//go:build tools

package deps

import (
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
	_ "github.com/vektra/mockery/v2" // Mockery is a tool for generating mocks for interfaces in Go. This prevents the tool from being removed when running go mod tidy.
)
