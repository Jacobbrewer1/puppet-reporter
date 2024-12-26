package api

import (
	"github.com/jacobbrewer1/pagefilter"
	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
)

type Repository interface {
	// GetReports gets all reports from the database
	GetReports(paginationDetails *pagefilter.PaginatorDetails, filters *GetReportsFilters) (*pagefilter.PaginatedResponse[models.Report], error)

	// GetReportByHash gets a report from the database by hash
	GetReportByHash(hash string) (*models.Report, error)

	// SaveReport saves a report to the database
	SaveReport(report *models.Report) error

	// SaveResources saves resources to the database
	SaveResources(resources []*models.Resource) error

	// SaveLogs saves logs to the database
	SaveLogs(logs []*models.LogMessage) error
}
