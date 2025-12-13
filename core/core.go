package core

import (
	"github.com/MoeclubM/V2bX/conf"
)

var coreFactory func(c *conf.CoreConfig) (Core, error)

func NewCore(c []conf.CoreConfig) (Core, error) {
	if len(c) == 0 {
		coreConfig := &conf.CoreConfig{}
		return coreFactory(coreConfig)
	}
	return coreFactory(&c[0])
}

func RegisterCore(f func(c *conf.CoreConfig) (Core, error)) {
	coreFactory = f
}
