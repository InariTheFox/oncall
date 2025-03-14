package server

import "github.com/InariTheFox/oncall/pkg/setting"

type Runner struct {
	Cfg *setting.Cfg
}

func NewRunner(cfg *setting.Cfg) Runner {
	return Runner{
		Cfg: cfg,
	}
}
