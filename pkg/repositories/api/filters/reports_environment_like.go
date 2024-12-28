package filters

import "github.com/jacobbrewer1/pagefilter"

type reportsEnvironmentLike struct {
	env string
}

func NewReportsEnvironmentLike(env string) pagefilter.Wherer {
	return &reportsEnvironmentLike{
		env: env,
	}
}

func (r *reportsEnvironmentLike) Where() (string, []any) {
	return "t.environment LIKE ?", []any{"%" + r.env + "%"}
}
