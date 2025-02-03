package filters

import "github.com/jacobbrewer1/pagefilter"

type reportsHostnameLike struct {
	hostname string
}

func NewReportHostnameLike(hostname string) pagefilter.Wherer {
	return &reportsHostnameLike{hostname: hostname}
}

func (f *reportsHostnameLike) Where() (string, []interface{}) {
	return "t.host LIKE ?", []interface{}{"%" + f.hostname + "%"}
}
