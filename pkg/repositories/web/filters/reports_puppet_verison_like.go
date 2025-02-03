package filters

import "github.com/jacobbrewer1/pagefilter"

type reportsPuppetVersionLike struct {
	puppetVersion string
}

func NewReportsPuppetVersionLike(puppetVersion string) pagefilter.Wherer {
	return &reportsPuppetVersionLike{puppetVersion: puppetVersion}
}

func (f *reportsPuppetVersionLike) Where() (string, []interface{}) {
	return "t.puppet_version LIKE ?", []interface{}{"%" + f.puppetVersion + "%"}
}
