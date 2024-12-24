package api

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/jacobbrewer1/puppet-reporter/pkg/logging"
	"github.com/jacobbrewer1/uhttp"
)

func (s *service) PostUpload(w http.ResponseWriter, r *http.Request) {
	l := logging.LoggerFromRequest(r)

	if r.Body == http.NoBody {
		l.Debug("No body in request")
		uhttp.SendMessageWithStatus(w, http.StatusBadRequest, "No body in request")
		return
	}

	bdy, err := io.ReadAll(r.Body)
	if err != nil {
		l.Error("Error reading body", slog.String("error", err.Error()))
		uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "Error reading body", err)
		return
	}

	rep, err := parsePuppetReport(bdy)
	if err != nil {
		l.Error("Error parsing puppet report", slog.String("error", err.Error()))
		uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "Error parsing puppet report", fmt.Errorf("error parsing puppet report: %w", err))
		return
	}

	l.Debug(fmt.Sprintf("Parsed puppet report: %v", rep))
}
