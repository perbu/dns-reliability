package config

import "time"

type Config struct {
	Interval       time.Duration `yaml:"interval"`
	ReportInterval time.Duration `yaml:"report_interval"`
	DNS            []DNS         `yaml:"dns"`
}
type Server struct {
	Name  string `yaml:"name"`
	Ipv4  string `yaml:"ipv4"`
	Ipv6  string `yaml:"ipv6"`
	Query string `yaml:"query"`
}
type DNS struct {
	Provider string   `yaml:"provider"`
	Servers  []Server `yaml:"servers"`
}
