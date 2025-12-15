package helpers

import "github.com/yetiz-org/gone/channel"

type ParamsHelper struct {
}

func (h *ParamsHelper) GetParam(params map[string]any, key channel.ParamKey) any {
	return params[string(key)]
}

func (h *ParamsHelper) SetParam(params map[string]any, key channel.ParamKey, value any) {
	params[string(key)] = value
}
