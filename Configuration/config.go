package Configuration

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type MysqlConfig struct {
	Addrs    string `yaml:"addrs,omitempty"`
	Port     string `yaml:"port,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Database string `yaml:"database,omitempty"`
	Table    string `yaml:"table,omitempty"`
	// configuration for sql connection pool
	ConnMaxLifetime int `yaml:"connMaxLifetime,omitempty"`
	MaxIdleConns    int `yaml:"maxIdleConns,omitempty"`
	MaxOpenConns    int `yaml:"maxOpenConns,omitempty"`
}

type Config struct {
	Taskconfig *MysqlConfig `yaml:"task,omitempty"`
}

// Validate the Mysql configuration
func validate_mysql(c *MysqlConfig) error {
	if c.Addrs == "" {
		return fmt.Errorf("mysql addrs is required")
	}
	if c.Port == "" {
		return fmt.Errorf("mysql port is required")
	}
	if c.Username == "" {
		return fmt.Errorf("mysql username is required")
	}
	if c.Password == "" {
		return fmt.Errorf("mysql password is required")
	}
	if c.Database == "" {
		return fmt.Errorf("mysql database is required")
	}
	if c.Table == "" {
		return fmt.Errorf("mysql table is required")
	}
	return nil
}

// validate the configuration
func validate(c *Config) error {
	if err := validate_mysql(c.Taskconfig); err != nil {
		return err
	}
	return nil
}

// LoadConfig parses configuration file and returns
// an initialized Settings object and an error object if any. For instance if it
// cannot find the configuration file it will set the returned error appropriately.
func LoadConfig(path string) (*Config, error) {
	c := &Config{}
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read configuration file: %s,error: %s", path, err)
	}
	if err = yaml.Unmarshal(contents, c); err != nil {
		return nil, fmt.Errorf("Failed to parse configuration,error: %s", err)
	}
	if err = validate(c); err != nil {
		return nil, fmt.Errorf("Invalid configuration,error: %s", err)
	}
	return c, nil
}
