package config

import (
	"bytes"
	"encoding/json"
	"html/template"
	"os"
	"reflect"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type SungrowConfig struct {
	Host     string `validate:"required"`
	Username string
	Password string
}

type Config struct {
	Host     string
	Port     uint
	LogLevel zerolog.Level
	Sungrow  SungrowConfig
}

func Load() (*Config, error) {
	v := viper.New()

	if err := loadDefaults(v); err != nil {
		return nil, err
	}

	if err := loadFromFiles(v, "redgiant",
		"/etc/redgiant",
		"$HOME/.config/redgiant",
		".",
	); err != nil {
		return nil, err
	}

	enableLoadFromEnvVars(v, "REDGIANT")

	c := &Config{}
	if err := v.Unmarshal(c, func(dc *mapstructure.DecoderConfig) {

		dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			stringTemplatingHookFunc(),
			stringToZerologLevelHookFunc(),
		)
	}); err != nil {
		return nil, err
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(c); err != nil {
		return nil, err
	}

	return c, nil
}

func loadDefaults(v *viper.Viper) error {
	dc := Config{
		Host:     "127.0.0.1",
		Port:     8000,
		LogLevel: zerolog.InfoLevel,
		Sungrow: SungrowConfig{
			Username: "user",
			Password: "pw1111",
		},
	}

	b, err := json.Marshal(dc)
	if err != nil {
		return err
	}

	r := bytes.NewReader(b)
	vv := viper.New()
	vv.SetConfigType("json")
	if err := vv.MergeConfig(r); err != nil {
		return err
	}

	v.MergeConfigMap(vv.AllSettings())
	return nil
}

func loadFromFiles(v *viper.Viper, configName string, paths ...string) error {
	for _, in := range paths {
		vv := viper.New()
		vv.SetConfigName(configName)
		vv.AddConfigPath(in)
		if err := vv.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return err
			}
		}

		v.MergeConfigMap(vv.AllSettings())
	}

	return nil
}

func enableLoadFromEnvVars(v *viper.Viper, prefix string) error {
	v.AutomaticEnv()
	v.SetEnvPrefix(prefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return nil
}

func stringTemplatingHookFunc() mapstructure.DecodeHookFuncType {
	e := map[string]string{}
	for _, kv := range os.Environ() {
		s := strings.SplitN(kv, "=", 2)
		e[s[0]] = s[1]
	}

	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		switch v := data.(type) {
		case string:
			tpl, err := template.New("").Funcs(sprig.FuncMap()).Parse(v)
			if err != nil {
				return nil, err
			}

			var b bytes.Buffer
			if err := tpl.Execute(&b, e); err != nil {
				return nil, err
			}

			return b.String(), nil
		default:
			return v, nil
		}
	}
}

func stringToZerologLevelHookFunc() mapstructure.DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(zerolog.NoLevel) {
			return data, nil
		}

		return zerolog.ParseLevel(data.(string))
	}
}
