package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

func ParseConfigBytes(data []byte) (Config, error) {
	var config Config
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, fmt.Errorf("yaml.Unmarshal: %w", err)
	}
	return config, nil
}

func ParseConfig(fh io.Reader) (Config, error) {
	content, err := io.ReadAll(fh)
	if err != nil {
		return Config{}, fmt.Errorf("io.ReadAll: %w", err)
	}
	return ParseConfigBytes(content)
}

func ParseConfigFile(filename string) (Config, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return Config{}, fmt.Errorf("os.Open: %w", err)
	}
	defer fh.Close()
	return ParseConfig(fh)
}
