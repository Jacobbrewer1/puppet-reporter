package api

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jacobbrewer1/pagefilter"
	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
	"github.com/jacobbrewer1/puppet-reporter/pkg/repositories/api/filters"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// ErrNoReports is returned when no reports are found.
	ErrNoReports = errors.New("no reports found")

	// ErrReportNotFound is returned when a report is not found.
	ErrReportNotFound = errors.New("report not found")
)

func (r *repository) SaveReport(report *models.Report) error {
	return report.Insert(r.db)
}

func (r *repository) GetReportByHash(hash string) (*models.Report, error) {
	rep, err := models.ReportByHash(r.db, hash)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrReportNotFound
		default:
			return nil, fmt.Errorf("get report by hash: %w", err)
		}
	}

	return rep, nil
}

func (r *repository) GetReports(paginationDetails *pagefilter.PaginatorDetails, filters *GetReportsFilters) (*pagefilter.PaginatedResponse[models.Report], error) {
	t := prometheus.NewTimer(models.DatabaseLatency.WithLabelValues("get_reports"))
	defer t.ObserveDuration()

	mf := r.getReportsFilters(filters)
	pg := pagefilter.NewPaginator(r.db, "report", "id", mf)

	if err := pg.SetDetails(paginationDetails, "id", "timestamp"); err != nil {
		return nil, fmt.Errorf("set paginator details: %w", err)
	}

	pvt, err := pg.Pivot()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoReports
		default:
			return nil, fmt.Errorf("set paginator details: %w", err)
		}
	}

	items := make([]*models.Report, 0)
	err = pg.Retrieve(pvt, &items)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoReports
		default:
			return nil, fmt.Errorf("failed to retrieve: %w", err)
		}
	}

	var total int64 = 0
	err = pg.Counts(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total: %w", err)
	}

	return &pagefilter.PaginatedResponse[models.Report]{
		Items: items,
		Total: total,
	}, nil
}

func (r *repository) getReportsFilters(f *GetReportsFilters) *pagefilter.MultiFilter {
	mf := pagefilter.NewMultiFilter()
	if f == nil {
		return mf
	}

	if f.Environment != nil {
		mf.Add(filters.NewReportsEnvironmentLike(*f.Environment))
	}

	if f.Host != nil {
		mf.Add(filters.NewReportsHostLike(*f.Host))
	}

	if f.State != nil {
		mf.Add(filters.NewReportsStateLike(*f.State))
	}

	return mf
}
