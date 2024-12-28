package filters

import "github.com/jacobbrewer1/pagefilter"

type reportsHostLike struct {
	host string
}

func NewReportsHostLike(host string) pagefilter.Wherer {
	return &reportsHostLike{
		host: host,
	}
}

func (r *reportsHostLike) Where() (string, []any) {
	return "t.host LIKE ?", []any{"%" + r.host + "%"}
}
