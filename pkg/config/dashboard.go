// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"github.com/kelseyhightower/envconfig"
)

// ChaosServerConfig defines the configuration for Chaos Dashboard
type ChaosServerConfig struct {
	ListenHost           string `envconfig:"LISTEN_HOST" default:"0.0.0.0"`
	ListenPort           int    `envconfig:"LISTEN_PORT" default:"2333"`
	MetricAddress        string `envconfig:"METRIC_ADDRESS"`
	EnableLeaderElection bool   `envconfig:"ENABLE_LEADER_ELECTION"`
	Database             *DatabaseConfig
}

// DatabaseConfig defines the configuration for databases
type DatabaseConfig struct {
	// Archive  Chaos Experiments to DB.
	Archive    bool
	Driver     string `envconfig:"DATABASE_DRIVER"     default:"sqlite3"`
	Datasource string `envconfig:"DATABASE_DATASOURCE" default:"core.sqlite"`
	Secret     string `envconfig:"DATABASE_SECRET"`
}

// EnvironChaosServer returns the settings from the environment.
func EnvironChaosServer() (ChaosServerConfig, error) {
	cfg := ChaosServerConfig{}
	err := envconfig.Process("", &cfg)
	return cfg, err
}
