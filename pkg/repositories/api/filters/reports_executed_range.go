package filters

import (
	"strings"
	"time"

	"github.com/jacobbrewer1/pagefilter"
)

type reportsExecutedRange struct {
	from time.Time
	to   time.Time
}

func NewReportsExecutedRange(from, to time.Time) pagefilter.Wherer {
	return &reportsExecutedRange{
		from: from,
		to:   to,
	}
}

func (r *reportsExecutedRange) Where() (string, []interface{}) {
	if !r.from.IsZero() && !r.to.IsZero() {
		return "t.executed_at BETWEEN ? AND ?", []interface{}{r.from, r.to}
	}

	builder := new(strings.Builder)
	args := make([]any, 0)

	if !r.from.IsZero() {
		builder.WriteString("t.executed_at >= ?")
		args = append(args, r.from)
	}

	if !r.to.IsZero() {
		if builder.Len() > 0 {
			builder.WriteString(" AND ")
		}
		builder.WriteString("t.executed_at <= ?")
		args = append(args, r.to)
	}

	return builder.String(), args
}
