package api

import (
	"github.com/jacobbrewer1/pagefilter"
	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
)

type Repository interface {
	// ListLatestHosts returns a list of the latest reports for each unique host.
	ListLatestHosts(details *pagefilter.PaginatorDetails, filters *ListLatestHostsFilters) (*pagefilter.PaginatedResponse[models.Report], error)
}

type ListLatestHostsFilters struct {
	// Hostname is the name of the host to filter by.
	Hostname *string

	// PuppetVersion is the version of puppet to filter by.
	PuppetVersion *string

	// Environment is the environment to filter by.
	Environment *string

	// Status is the status to filter by.
	Status *string
}
