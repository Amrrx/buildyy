// pkg/config/config.go
package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type SubProject struct {
	Name       string   `yaml:"name"`
	Version    string   `yaml:"version"`
	Path       string   `yaml:"path"`
	BuildCmd   []string `yaml:"buildCmd"`
	Dockerfile string   `yaml:"dockerfile"`
	DependsOn  []string `yaml:"dependsOn"`
}

type Config struct {
	SubProjects []SubProject `yaml:"subProjects"`
}

func ParseConfig(configFile string) (*Config, error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return &config, nil
}
