package filters

import "github.com/jacobbrewer1/pagefilter"

type reportsStateLike struct {
	state string
}

func NewReportsStateLike(state string) pagefilter.Wherer {
	return &reportsStateLike{state: state}
}

func (f *reportsStateLike) Where() (string, []interface{}) {
	return "t.state LIKE ?", []interface{}{"%" + f.state + "%"}
}
