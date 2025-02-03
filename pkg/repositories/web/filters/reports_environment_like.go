package filters

import "github.com/jacobbrewer1/pagefilter"

type reportsEnvironmentLike struct {
	environment string
}

func NewReportsEnvironmentLike(environment string) pagefilter.Wherer {
	return &reportsEnvironmentLike{environment: environment}
}

func (f *reportsEnvironmentLike) Where() (string, []interface{}) {
	return "t.environment LIKE ?", []interface{}{"%" + f.environment + "%"}
}
