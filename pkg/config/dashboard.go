// Copyright 2020 Chaos Mesh Authors.
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

// ChaosDashboardConfig defines the configuration for Chaos Dashboard
type ChaosDashboardConfig struct {
	ListenHost           string `envconfig:"LISTEN_HOST" default:"0.0.0.0"`
	ListenPort           int    `envconfig:"LISTEN_PORT" default:"2333"`
	MetricAddress        string `envconfig:"METRIC_ADDRESS"`
	EnableLeaderElection bool   `envconfig:"ENABLE_LEADER_ELECTION"`
	Database             *DatabaseConfig
	PersistTTL           PersistTTLConfig
}

// PersistTTLConfig defines the configuration of ttl
type PersistTTLConfig struct {
	SyncPeriod string `envconfig:"CLEAN_SYNC_PERIOD" default:"12h"`
	Event      string `envconfig:"TTL_EVENT"       default:"168h"`
	Experiment string `envconfig:"TTL_EXPERIMENT"  default:"336h"`
}

// DatabaseConfig defines the configuration for databases
type DatabaseConfig struct {
	// Archive  Chaos Experiments to DB.
	Archive    bool
	Driver     string `envconfig:"DATABASE_DRIVER"     default:"sqlite3"`
	Datasource string `envconfig:"DATABASE_DATASOURCE" default:"core.sqlite"`
	Secret     string `envconfig:"DATABASE_SECRET"`
}

// EnvironChaosDashboard returns the settings from the environment.
func EnvironChaosDashboard() (ChaosDashboardConfig, error) {
	cfg := ChaosDashboardConfig{}
	err := envconfig.Process("", &cfg)
	return cfg, err
}
