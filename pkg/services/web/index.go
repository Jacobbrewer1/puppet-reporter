package web

import (
	"html/template"
	"net/http"

	"github.com/jacobbrewer1/uhttp"
)

func (s *service) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("index").ParseFS(templates, "templates/index.gohtml"))

	if err := tmpl.Execute(w, nil); err != nil {
		uhttp.SendErrorMessageWithStatus(w, http.StatusInternalServerError, "Error executing template", err)
		return
	}
}
