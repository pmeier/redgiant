package cmd

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Viper struct {
	*viper.Viper
}

func NewViper() *Viper {
	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvPrefix("redgiant")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	return &Viper{Viper: v}
}

func decoderConfigOption(dc *mapstructure.DecoderConfig) {
	dc.MatchName = matchName
}

func matchName(mapKey string, fieldName string) bool {
	return strings.EqualFold(strings.Replace(mapKey, "-", "", -1), fieldName)
}

func (v *Viper) Unmarshal(rawVal any) error {
	return v.Viper.Unmarshal(rawVal, decoderConfigOption)
}
