package api

import "github.com/jacobbrewer1/puppet-reporter/pkg/models"

func (r *repository) SaveReport(report *models.Report) error {
	return report.Insert(r.db)
}

func (r *repository) GetReportByHash(hash string) (*models.Report, error) {
	return models.ReportByHash(r.db, hash)
}
