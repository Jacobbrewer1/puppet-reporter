package web

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jacobbrewer1/puppet-reporter/pkg/logging"
	"github.com/jacobbrewer1/puppet-reporter/pkg/models"
	repo "github.com/jacobbrewer1/puppet-reporter/pkg/repositories/web"
	"github.com/jacobbrewer1/uhttp"
)

const (
	indexTemplateReportListName = "report-list-element"
)

func (s *service) getReportListHandler(w http.ResponseWriter, r *http.Request) {
	previousDaysCountStr := r.PostFormValue("num-days")
	if previousDaysCountStr == "" {
		uhttp.SendErrorMessageWithStatus(w, http.StatusBadRequest, "num-days not provided", nil)
		return
	}

	previousDaysCount, err := strconv.Atoi(previousDaysCountStr)
	if err != nil {
		uhttp.SendErrorMessageWithStatus(w, http.StatusBadRequest, "num-days must be an integer", err)
		return
	}

	now := time.Now().UTC()

	// Set now to the end of the day
	now = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)

	start := now.Add(-time.Hour * 24 * time.Duration(previousDaysCount))

	// Set start to the beginning of the day
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)

	reps, err := s.r.GetReportsInPeriod(start, now)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNoReports):
			reps = make([]*models.Report, 0)
		default:
			slog.Error("Error getting reports", slog.String(logging.KeyError, err.Error()))
			uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "Error getting reports", err)
			return
		}
	}

	tmpl := template.Must(template.New("index").Funcs(
		template.FuncMap{
			"getReportStyle": getReportStyle,
		}).ParseFS(templates, "templates/index.gohtml"))

	tmplTpe := struct {
		Reports []*models.Report
	}{
		Reports: reps,
	}

	if err := tmpl.ExecuteTemplate(w, indexTemplateReportListName, tmplTpe); err != nil {
		slog.Error("Error executing template", slog.String(logging.KeyError, err.Error()))
		uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "Error executing template", err)
		return
	}
}
