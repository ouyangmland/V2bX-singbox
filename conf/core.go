package conf

import (
	"encoding/json"
)

type CoreConfig struct {
	Name       string      `json:"Name"`
	SingConfig *SingConfig `json:"-"`
}

func (c *CoreConfig) UnmarshalJSON(b []byte) error {
	type _CoreConfig CoreConfig
	err := json.Unmarshal(b, (*_CoreConfig)(c))
	if err != nil {
		return err
	}
	c.SingConfig = NewSingConfig()
	return json.Unmarshal(b, c.SingConfig)
}
