package main

import (
	"errors"
	"fmt"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"os"
)

// Config object
type Config struct {
	Common struct {
		Configpath string
		Logpath    string
	}

	Directory struct {
		User     string
		Password string
		Host     string
		Port     int64

		Groups   map[string][]string
		Roles    map[string][]string
		Attrmap  map[string]map[string]string
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

// NewConfig creates new object instance
func NewConfig() *Config {
	cfg := new(Config)
	cfg.Directory.Groups = make(map[string][]string)
	cfg.Directory.Roles = make(map[string][]string)
	cfg.Directory.Attrmap = make(map[string]map[string]string)

	return cfg
}

// ConfigReader object
type ConfigReader struct {
	path   string
	config *Config
}

// NewConfigReader creates new object instance
func NewConfigReader(path string) *ConfigReader {
	cfg := new(ConfigReader)
	cfg.path = path
	cfg.config = NewConfig()
	cfg.loadFromPath()

	return cfg.validate()
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
	} else {
		cfg.setDefaults()
	}
}

// Set defaults if they were not configured
func (cfg *ConfigReader) setDefaults() {
	if cfg.Config().Common.Configpath == "" {
		cfg.config.Common.Configpath = "/etc/rhn/ldapsync.conf"
	}

	if cfg.Config().Common.Logpath == "" {
		cfg.config.Common.Logpath = "/var/log/rhn/ldapsync.log"
	}

	if cfg.Config().Directory.Port == 0 {
		cfg.config.Directory.Port = 389
	}
}

func (cfg *ConfigReader) validateAggregate(aggr map[string][]string) error {
	if len(aggr) == 0 {
		return errors.New("Block is empty")
	}

	for dn := range aggr {
		if len(dn) == 0 {
			return fmt.Errorf("DN '%s' contains no mapped roles", dn)
		}
	}

	return nil
}

// Validate the configuration, if it is eligible to proceed with the syncing
func (cfg *ConfigReader) validate() *ConfigReader {
	for errmsg, attr := range map[string]interface{}{
		// Directory
		"DN for LDAP user is not specified":                            cfg.config.Directory.User,
		"Password for LDAP user is not specified":                      cfg.config.Directory.Password,
		"Fully qualified domain name for LDAP server is not specified": cfg.config.Directory.Host,
		"DN for all LDAP users is not specified":                       cfg.config.Directory.Allusers,

		// Uyuni
		"Uyuni RPC-API URL is not specified":               cfg.config.Spacewalk.Url,
		"Uyuni user is not specified":                      cfg.config.Spacewalk.User,
		"The password for the Uyuni user is not specified": cfg.config.Spacewalk.Password} {
		if attr == "" {
			log.Fatal(errmsg)
		}
	}

	// Look if at least one frozen dude has this role
	if len(cfg.config.Directory.Frozen) == 0 {
		log.Fatal("You have to regiser at least one frozen account with Organisation Manager role for emergency purposes")
	}

	// Look if at least one frozen dude has this role
	if len(cfg.config.Directory.Groups) == 0 && len(cfg.config.Directory.Roles) == 0 {
		log.Fatal("Either Directory/Groups or Directory/Roles needs to be specified")
	}

	for _, aggr := range []map[string][]string{cfg.config.Directory.Groups, cfg.config.Directory.Roles} {
		err := cfg.validateAggregate(aggr)
		if err != nil {
			log.Fatal(err)
		}
	}

	return cfg
}

// Config returns the configuration object
func (cfg *ConfigReader) Config() *Config {
	return cfg.config
}
