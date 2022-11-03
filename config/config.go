package config

import (
	"gopkg.in/yaml.v3"
	"io"
)

type (
	Config struct {
		//
		Api struct {
			//
			Port uint16 `yaml:"port"`
			//
			Matchers Matchers `yaml:"matchers"`
		} `yaml:"api"`
		//
		Db Db `yaml:"db"`
	}

	Db struct {
		//
		Uri string `yaml:"uri"`
		//
		Name string `yaml:"name"`
		//
		Table struct {
			//
			Name string `yaml:"name"`
		}
	}

	Matchers struct {
		//
		Uri string `yaml:"uri"`
	}
)

func NewConfigFromYaml(r io.Reader) (cfg Config, err error) {
	err = yaml.NewDecoder(r).Decode(&cfg)
	return
}
