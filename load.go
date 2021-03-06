package multiconfig

import (
	"flag"
	"strings"

	"github.com/arstd/log"
)

// Confer get flags' configuration file, if not blank, load this configuration file
// after load in turn, then load flags again.
type Confer interface {
	GetConf() string
}

type Conf struct {
	Conf     string `flagUsage:"conf file, it will override other conf file but not flags"`
	Name     string `default:"aclogs" flagUsage:"server name"`
	LogLevel string `default:"info" flagUsage:"log level, trace/debug/info/warn/error/fatal"`
}

func (c *Conf) GetConf() string { return c.Conf }

// LoadInTurn load configuration from tag/json/yml/yaml/toml/env/flag(from low to heigh) and validate.
// Configuration file in $PWD/conf/$app.{json,yml,yaml/toml}.
func LoadInTurn(conf Confer) error {
	flagLoader := &FlagLoader{
		Prefix:        "",
		Flatten:       false,
		CamelCase:     true,
		ErrorHandling: flag.ExitOnError,
	}

	for _, loader := range []Loader{
		&TagLoader{},
		&JSONLoader{Path: "conf/conf.json"},
		&YAMLLoader{Path: "conf/conf.yml"},
		&YAMLLoader{Path: "conf/conf.yaml"},
		&TOMLLoader{Path: "conf/conf.toml"},
		&EnvironmentLoader{CamelCase: true},
		flagLoader,
	} {
		if err := loader.Load(conf); err != nil {
			log.Warn(err)
		}
	}

	if conf.GetConf() != "" {
		path := conf.GetConf()
		loaders := []Loader{}
		if strings.HasSuffix(path, "toml") {
			loaders = append(loaders, &TOMLLoader{Path: path})
		} else if strings.HasSuffix(path, "json") {
			loaders = append(loaders, &JSONLoader{Path: path})
		} else if strings.HasSuffix(path, "yml") || strings.HasSuffix(path, "yaml") {
			loaders = append(loaders, &YAMLLoader{Path: path})
		}
		loaders = append(loaders, &FlagLoader{})
		loader := MultiLoader(loaders...)
		if err := loader.Load(conf); err != nil {
			return err
		}
	}

	return (&RequiredValidator{}).Validate(conf)
}
