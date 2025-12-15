package helpers

import "github.com/yetiz-org/goth-scaffold/app/conf"

type ParamsConfigHelper struct {
	ParamsHelper
}

func (h *ParamsConfigHelper) Config() *conf.Configuration {
	return conf.Config()
}
