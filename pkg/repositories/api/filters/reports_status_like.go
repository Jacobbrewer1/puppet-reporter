package filters

import "github.com/jacobbrewer1/pagefilter"

type reportsStatusLike struct {
	status string
}

func NewReportsStatusLike(status string) pagefilter.Wherer {
	return &reportsStatusLike{
		status: status,
	}
}

func (r *reportsStatusLike) Where() (string, []any) {
	return "t.state LIKE ?", []any{"%" + r.status + "%"}
}
