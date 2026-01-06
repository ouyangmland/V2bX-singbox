package core

import (
	"fmt"
	"github.com/MoeclubM/V2bX/conf"
)

var coreFactory func(c *conf.CoreConfig) (Core, error)

func NewCore(c []conf.CoreConfig) (Core, error) {
	if coreFactory == nil {
		return nil, fmt.Errorf("no core registered (coreFactory is nil)")
	}
	if len(c) == 0 {
		return nil, fmt.Errorf("no core config found: please configure at least one item in `Cores`")
	}
	return coreFactory(&c[0])
}

func RegisterCore(f func(c *conf.CoreConfig) (Core, error)) {
	coreFactory = f
}
