package api

import (
	"github.com/jacobbrewer1/puppet-reporter/pkg/apis/specs/api"
	repo "github.com/jacobbrewer1/puppet-reporter/pkg/repositories/api"
)

type service struct {
	// r is the repository used by the service.
	r repo.Repository
}

func NewService(r repo.Repository) api.ServerInterface {
	return &service{
		r: r,
	}
}
