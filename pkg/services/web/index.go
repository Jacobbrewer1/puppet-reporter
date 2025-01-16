package web

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/jacobbrewer1/puppet-reporter/pkg/logging"
	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
	repo "github.com/jacobbrewer1/puppet-reporter/pkg/repositories/web"
	"github.com/jacobbrewer1/uhttp"
)

const (
	defaultReportSearchRange = 7
)

func (s *service) indexHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)

	start := now.Add(-time.Hour * 24 * time.Duration(defaultReportSearchRange))
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)

	reps, err := s.r.GetLatestUniqueReportHosts(start, now)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNoReports):
			reps = make([]*models.Report, 0)
		default:
			slog.Error("Error getting reports", slog.String(logging.KeyError, err.Error()))
			uhttp.SendMessageWithStatus(w, http.StatusInternalServerError, "error getting reports")
			return
		}
	}

	tmpl := template.Must(template.New("index").Funcs(
		template.FuncMap{
			"getReportStyle": getReportStyle,
		}).ParseFS(templates, "templates/index.gohtml"))

	tmplTpe := struct {
		Reports                   []*models.Report
		ReportsDefaultSearchRange int
	}{
		Reports:                   reps,
		ReportsDefaultSearchRange: defaultReportSearchRange,
	}

	if err := tmpl.Execute(w, tmplTpe); err != nil {
		uhttp.SendMessageWithStatus(w, http.StatusInternalServerError, "error rendering template")
		return
	}
}
