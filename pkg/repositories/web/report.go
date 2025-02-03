package api

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jacobbrewer1/pagefilter"
	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
	"github.com/jacobbrewer1/puppet-reporter/pkg/repositories/web/filters"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// ErrNoReports is returned when no reports are found.
	ErrNoReports = errors.New("no reports found")
)

func (r *repository) ListLatestHosts(details *pagefilter.PaginatorDetails, filters *ListLatestHostsFilters) (*pagefilter.PaginatedResponse[models.Report], error) {
	t := prometheus.NewTimer(models.DatabaseLatency.WithLabelValues("list_latest_hosts"))
	defer t.ObserveDuration()

	mf := r.getListLatestHostFilters(filters)
	pg := pagefilter.NewPaginator(r.db, models.ReportTableName, "id", mf)

	if err := pg.SetDetails(details, "host", "executed_at"); err != nil {
		return nil, fmt.Errorf("set paginator details: %w", err)
	}

	pvt, err := pg.Pivot()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoReports
		default:
			return nil, fmt.Errorf("paginate reports: %w", err)
		}
	}

	reports := make([]*models.Report, 0)
	if err := pg.Retrieve(pvt, &reports); err != nil {
		return nil, fmt.Errorf("retrieve reports: %w", err)
	}

	var total int64 = 0
	if err := pg.Counts(&total); err != nil {
		return nil, fmt.Errorf("get total count: %w", err)
	}

	return &pagefilter.PaginatedResponse[models.Report]{
		Items: reports,
		Total: total,
	}, nil
}

func (r *repository) getListLatestHostFilters(f *ListLatestHostsFilters) *pagefilter.MultiFilter {
	mf := pagefilter.NewMultiFilter()
	if f == nil {
		return mf
	}

	if f.Hostname != nil {
		mf.Add(filters.NewReportHostnameLike(*f.Hostname))
	}

	if f.PuppetVersion != nil {
		mf.Add(filters.NewReportsPuppetVersionLike(*f.PuppetVersion))
	}

	if f.Environment != nil {
		mf.Add(filters.NewReportsEnvironmentLike(*f.Environment))
	}

	if f.Status != nil {
		mf.Add(filters.NewReportsStateLike(*f.Status))
	}

	return mf
}
