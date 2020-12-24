package config

import (
	"strings"

	"github.com/deltacat/dbstress/utils"
)

// FindDefaultMySQLConnection find default configuration from configured connections
func (c *Config) FindDefaultMySQLConnection() (MySQLClientConfig, error) {
	v := c.Connection.MySQL
	for _, sc := range v {
		if sc.Default {
			return sc, nil
		}
	}
	if len(v) > 0 {
		return v[0], nil
	}
	return MySQLClientConfig{}, utils.ErrNotFound
}

// FindDefaultInfluxDBConnection find default configuration from configured connections
func (c *Config) FindDefaultInfluxDBConnection() (InfluxClientConfig, error) {
	v := c.Connection.InfluxDB
	for _, sc := range v {
		if sc.Default {
			return sc, nil
		}
	}
	if len(v) > 0 {
		return v[0], nil
	}
	return InfluxClientConfig{}, utils.ErrNotFound

}

// FindMySQLConnection find connnection by name
func (c *Config) FindMySQLConnection(name string) (MySQLClientConfig, error) {
	v := c.Connection.MySQL
	for _, sc := range v {
		if strings.EqualFold(sc.Name, name) {
			return sc, nil
		}
	}
	return MySQLClientConfig{}, nil
}

// FindInfluxDBConnection find connnection by name
func (c *Config) FindInfluxDBConnection(name string) (InfluxClientConfig, error) {
	v := c.Connection.InfluxDB
	for _, sc := range v {
		if sc.Default {
			return sc, nil
		}
	}
	return InfluxClientConfig{}, nil
}
