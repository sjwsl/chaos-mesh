// Copyright 2019 PingCAP, Inc.
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

package main

import (
	"flag"
	"os"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/pingcap/chaos-mesh/pkg/chaosdaemon"
	"github.com/pingcap/chaos-mesh/pkg/version"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	log  = ctrl.Log.WithName("chaos-daemon")
	conf = &chaosdaemon.Config{Host: "0.0.0.0"}

	printVersion bool
)

func init() {
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")
	flag.IntVar(&conf.GRPCPort, "grpc-port", 31767, "the port which grpc server listens on")
	flag.IntVar(&conf.HTTPPort, "http-port", 31766, "the port which http server listens on")
	flag.StringVar(&conf.Runtime, "runtime", "docker", "current container runtime")
	flag.BoolVar(&conf.Profiling, "pprof", false, "enable pprof")

	flag.Parse()
}

func main() {
	version.PrintVersionInfo("Chaos-daemon")

	if printVersion {
		os.Exit(0)
	}

	ctrl.SetLogger(zap.Logger(true))

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)

	if err := chaosdaemon.StartServer(conf, reg); err != nil {
		log.Error(err, "failed to start chaos-daemon server")
		os.Exit(1)
	}
}
