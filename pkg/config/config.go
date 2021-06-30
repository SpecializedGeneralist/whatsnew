// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

// Config holds whatsnew application-wide configuration settings.
type Config struct {
	DB DB `yaml:"db"`
}

// DB holds database settings.
type DB struct {
	// DSN, dbname excluded.
	DSN    string `yaml:"dsn"`
	DBName string `yaml:"dbname"`
}

// FromYAMLFile reads a Config object from a YAML file.
//
// Before being decoded, the whole YAML file content is passed through
// os.ExpandEnv.
func FromYAMLFile(filename string) (*Config, error) {
	rawContent, err := os.ReadFile(filename)
	if err != nil {
		err = fmt.Errorf("cannot read config file %#v: %w", filename, err)
		return nil, err
	}
	content := os.ExpandEnv(string(rawContent))

	conf := new(Config)
	err = yaml.NewDecoder(strings.NewReader(content)).Decode(conf)
	if err != nil {
		err = fmt.Errorf("cannot decode config file %#v: %w", filename, err)
		return nil, err
	}
	return conf, nil
}
