package main

import (
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Directory struct {
		User     string
		Password string
		Host     string
		Port     int64

		Groups   map[string][]string
		Roles    map[string][]string
		Frozen   []string
		Allusers string
	}
	Spacewalk struct {
		Url      string
		User     string
		Password string
		Checkssl bool
	}
}

func NewConfig() *Config {
	cfg := new(Config)
	cfg.Directory.Groups = make(map[string][]string)
	cfg.Directory.Roles = make(map[string][]string)

	return cfg
}

type ConfigReader struct {
	path   string
	config *Config
}

func NewConfigReader(path string) *ConfigReader {
	cfg := new(ConfigReader)
	cfg.path = path
	cfg.config = NewConfig()
	cfg.loadFromPath()

	return cfg
}

// Load configuration from the path
func (cfg *ConfigReader) loadFromPath() {
	fh, err := os.Open(cfg.path)
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()
	cfgBytes, err := ioutil.ReadAll(fh)
	if err != nil {
		log.Fatal(err)
	}

	if err := yaml.Unmarshal(cfgBytes, &cfg.config); err != nil {
		log.Fatal(err)
	}
}

func (cfg *ConfigReader) Config() *Config {
	return cfg.config
}
