package api

import (
	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
)

type Repository interface {
	// SaveReport saves a report to the database
	SaveReport(report *models.Report) error
}
