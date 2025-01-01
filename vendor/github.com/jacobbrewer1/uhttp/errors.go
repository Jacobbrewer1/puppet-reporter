package uhttp

import "net/http"

type StatusCoder interface {
	StatusCode() int
}

type HTTPError struct {
	error

	Title   string        `json:"title"`
	Detail  string        `json:"detail"`
	Status  int           `json:"status"`
	Details []interface{} `json:"details,omitempty"`

	// RequestID is our addition in order to be able to trace requests in the log.
	RequestID string `json:"request_id,omitempty"`
}

func (e *HTTPError) Error() string {
	return e.error.Error()
}

func (e *HTTPError) StatusCode() int {
	if e.Status == 0 {
		return http.StatusOK
	}
	return e.Status
}

// NewHTTPError creates a new HTTPError.
func NewHTTPError(code int, err error, details ...any) *HTTPError {
	return &HTTPError{
		error:   err,
		Title:   http.StatusText(code),
		Detail:  err.Error(),
		Details: details,
		Status:  code,

		// RequestID will be populated at error write time
		RequestID: "",
	}
}
