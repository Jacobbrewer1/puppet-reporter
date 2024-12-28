package filters

import "github.com/jacobbrewer1/pagefilter"

type reportsStateLike struct {
	state string
}

func NewReportsStateLike(state string) pagefilter.Wherer {
	return &reportsStateLike{
		state: state,
	}
}

func (r *reportsStateLike) Where() (string, []any) {
	return "t.state LIKE ?", []any{"%" + r.state + "%"}
}
