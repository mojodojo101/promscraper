package config

import (
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config represents the configuration for the exporter
type Config struct {
	Sql sql `yaml:"sql,omitempty"`
	Prometheus prometheus `yaml:"prometheus,omitempty"`
}

type sql struct {
	Name     string `yaml:"name,omitempty"`
	Address  string `yaml:"address,omitempty"`
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type prometheus struct {
	Endpoint string `yaml:"endpoint,omitempty"`
}

func New() *Config {
	c := &Config{}
	return c
}

// Load loads a config from reader
func Load(reader io.Reader) (*Config, error) {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	c := New()
	err = yaml.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
