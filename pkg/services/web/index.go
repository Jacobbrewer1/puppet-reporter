package web

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/jacobbrewer1/pagefilter"
	"github.com/jacobbrewer1/puppet-reporter/pkg/logging"
	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
	repo "github.com/jacobbrewer1/puppet-reporter/pkg/repositories/web"
	"github.com/jacobbrewer1/uhttp"
)

const (
	defaultReportSearchRange = 15
)

func (s *service) indexHandler(w http.ResponseWriter, r *http.Request) {
	details, err := pagefilter.DetailsFromRequest(r)
	if err != nil {
		slog.Error("Error getting page details", slog.String(logging.KeyError, err.Error()))
		uhttp.SendMessageWithStatus(w, http.StatusBadRequest, "invalid page details")
		return
	}

	if r.URL.Query().Get("limit") == "" {
		details.Limit = defaultReportSearchRange
	}
	if details.SortBy == "" {
		details.SortBy = "executed_at"
	}
	if details.SortDir == "" {
		details.SortDir = "desc" // Latest reports first
	}

	filters := s.getListReportFilters(
		r.URL.Query().Get("host"),
		r.URL.Query().Get("puppet_version"),
		r.URL.Query().Get("environment"),
		r.URL.Query().Get("status"),
	)

	reps, err := s.r.ListLatestHosts(details, filters)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNoReports):
			reps = &pagefilter.PaginatedResponse[models.Report]{
				Items: make([]*models.Report, 0),
				Total: 0,
			}
		default:
			slog.Error("Error getting reports", slog.String(logging.KeyError, err.Error()))
			uhttp.SendMessageWithStatus(w, http.StatusInternalServerError, "error getting reports")
			return
		}
	}

	tmpl := template.Must(template.New("index").Funcs(
		template.FuncMap{
			"getReportStyle": getReportStyle,
		},
	).ParseFS(localTemplates, "templates/index.gohtml"))

	tmplTpe := struct {
		Reports *pagefilter.PaginatedResponse[models.Report]
	}{
		Reports: reps,
	}

	if err := tmpl.Execute(w, tmplTpe); err != nil {
		slog.Error("Error rendering template", slog.String(logging.KeyError, err.Error()))
		uhttp.SendMessageWithStatus(w, http.StatusInternalServerError, "error rendering template")
		return
	}
}

func (s *service) APIListReports(w http.ResponseWriter, r *http.Request) {
	details, err := pagefilter.DetailsFromRequest(r)
	if err != nil {
		uhttp.SendMessageWithStatus(w, http.StatusBadRequest, "invalid page details")
		return
	}

	if r.URL.Query().Get("limit") == "" {
		details.Limit = defaultReportSearchRange
	}
	if details.SortBy == "" {
		details.SortBy = "executed_at"
	}
	if details.SortDir == "" {
		details.SortDir = "desc" // Latest reports first
	}

	filters := s.getListReportFilters(
		r.URL.Query().Get("host"),
		r.URL.Query().Get("puppet-version"),
		r.URL.Query().Get("environment"),
		r.URL.Query().Get("status"),
	)

	reps, err := s.r.ListLatestHosts(details, filters)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNoReports):
			reps = &pagefilter.PaginatedResponse[models.Report]{
				Items: make([]*models.Report, 0),
				Total: 0,
			}
		default:
			uhttp.SendMessageWithStatus(w, http.StatusInternalServerError, "error getting reports")
			return
		}
	}

	tmpl := template.Must(template.New("index").Funcs(
		template.FuncMap{
			"getReportStyle": getReportStyle,
		},
	).ParseFS(localTemplates, "templates/index.gohtml"))

	tmplTpe := struct {
		Reports *pagefilter.PaginatedResponse[models.Report]
	}{
		Reports: reps,
	}

	if err := tmpl.ExecuteTemplate(w, "report_list", tmplTpe); err != nil {
		uhttp.SendMessageWithStatus(w, http.StatusInternalServerError, "error rendering template")
		return
	}
}

func (s *service) APIReportsTotal(w http.ResponseWriter, r *http.Request) {
	details, err := pagefilter.DetailsFromRequest(r)
	if err != nil {
		uhttp.SendMessageWithStatus(w, http.StatusBadRequest, "invalid page details")
		return
	}

	if r.URL.Query().Get("limit") == "" {
		details.Limit = defaultReportSearchRange
	}
	if details.SortBy == "" {
		details.SortBy = "executed_at"
	}
	if details.SortDir == "" {
		details.SortDir = "desc" // Latest reports first
	}

	filters := s.getListReportFilters(
		r.URL.Query().Get("host"),
		r.URL.Query().Get("puppet_version"),
		r.URL.Query().Get("environment"),
		r.URL.Query().Get("status"),
	)

	reps, err := s.r.ListLatestHosts(details, filters)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNoReports):
			reps = &pagefilter.PaginatedResponse[models.Report]{
				Items: make([]*models.Report, 0),
				Total: 0,
			}
		default:
			uhttp.SendMessageWithStatus(w, http.StatusInternalServerError, "error getting reports")
			return
		}
	}

	tmpl := template.Must(template.New("index").Funcs(
		template.FuncMap{
			"getReportStyle": getReportStyle,
		},
	).ParseFS(localTemplates, "templates/index.gohtml"))

	tmplTpe := struct {
		Reports *pagefilter.PaginatedResponse[models.Report]
	}{
		Reports: reps,
	}

	if err := tmpl.ExecuteTemplate(w, "total_reports", tmplTpe); err != nil {
		uhttp.SendMessageWithStatus(w, http.StatusInternalServerError, "error rendering template")
		return
	}
}

func (s *service) getListReportFilters(
	host string,
	puppetVersion string,
	environment string,
	status string,
) *repo.ListLatestHostsFilters {
	filters := new(repo.ListLatestHostsFilters)
	if host != "" {
		filters.Hostname = &host
	}

	if puppetVersion != "" {
		filters.PuppetVersion = &puppetVersion
	}

	if environment != "" {
		filters.Environment = &environment
	}

	if status != "" {
		filters.Status = &status
	}

	return filters
}
