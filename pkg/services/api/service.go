package api

import (
	"github.com/jacobbrewer1/puppet-reporter/pkg/codegen/apis/api"
)

type service struct{}

func NewService() api.ServerInterface {
	return &service{}
}
