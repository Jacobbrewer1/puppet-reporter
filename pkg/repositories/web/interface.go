package api

import (
	"time"

	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
)

type Repository interface {
	// GetLatestUniqueReportHosts returns a list of the latest reports for each unique host.
	GetLatestUniqueReportHosts(start, end time.Time) ([]*models.Report, error)

	// GetReportsInPeriod returns a list of reports within the given period.
	GetReportsInPeriod(start, end time.Time) ([]*models.Report, error)
}
