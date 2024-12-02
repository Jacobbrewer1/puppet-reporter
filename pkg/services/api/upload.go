package api

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/jacobbrewer1/uhttp"
)

func (s *service) PostUpload(w http.ResponseWriter, r *http.Request) {
	if r.Body == http.NoBody {
		slog.Debug("No body in request")
		uhttp.SendMessageWithStatus(w, http.StatusBadRequest, "No body in request")
		return
	}

	bdy, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Error reading body", slog.String("error", err.Error()))
		uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "Error reading body", err)
		return
	}

	rep, err := parsePuppetReport(bdy)
	if err != nil {
		slog.Error("Error parsing puppet report", slog.String("error", err.Error()))
		uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "Error parsing puppet report", fmt.Errorf("error parsing puppet report: %w", err))
		return
	}

	slog.Debug(fmt.Sprintf("Parsed puppet report: %v", rep))
}
