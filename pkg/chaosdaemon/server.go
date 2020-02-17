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

package chaosdaemon

import (
	"context"
	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	ctrl "sigs.k8s.io/controller-runtime"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/utils"
)

var log = ctrl.Log.WithName("chaos-daemon-server")

//go:generate protoc -I pb pb/chaosdaemon.proto --go_out=plugins=grpc:pb

// Config contains the basic chaos daemon configuration.
type Config struct {
	HTTPPort  int
	GRPCPort  int
	Host      string
	Runtime   string
	Profiling bool
}

// Server represents a grpc server for tc daemon
type daemonServer struct {
	crClient ContainerRuntimeInfoClient
}

func newDaemonServer(containerRuntime string) (*daemonServer, error) {
	crClient, err := CreateContainerRuntimeInfoClient(containerRuntime)
	if err != nil {
		return nil, err
	}

	return &daemonServer{
		crClient: crClient,
	}, nil
}

func newGRPCServer(containerRuntime string, reg *prometheus.Registry) (*grpc.Server, error) {
	ds, err := newDaemonServer(containerRuntime)
	if err != nil {
		return nil, err
	}

	grpcMetrics := grpc_prometheus.NewServerMetrics()
	grpcMetrics.EnableHandlingTimeHistogram(
		grpc_prometheus.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 10}),
	)
	reg.MustRegister(grpcMetrics)

	grpcOpts := []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			utils.TimeoutServerInterceptor,
			grpcMetrics.UnaryServerInterceptor(),
		),
	}

	s := grpc.NewServer(grpcOpts...)
	grpcMetrics.InitializeMetrics(s)

	pb.RegisterChaosDaemonServer(s, ds)
	reflection.Register(s)

	return s, nil
}

// StartServer starts chaos-daemon.
func StartServer(conf *Config, reg *prometheus.Registry) error {
	g := errgroup.Group{}

	httpBindAddr := fmt.Sprintf("%s:%d", conf.Host, conf.HTTPPort)
	srv := newHTTPServer(httpBindAddr, conf.Profiling, reg)
	g.Go(func() error {
		log.Info("starting http endpoint", "address", httpBindAddr)
		if err := srv.ListenAndServe(); err != nil {
			log.Error(err, "failed to start http endpoint")
			srv.Shutdown(context.Background())
			return err
		}
		return nil
	})

	grpcServer, err := newGRPCServer(conf.Runtime, reg)
	if err != nil {
		log.Error(err, "failed to create grpc server")
		return err
	}

	grpcBindAddr := fmt.Sprintf("%s:%d", conf.Host, conf.GRPCPort)
	lis, err := net.Listen("tcp", grpcBindAddr)
	if err != nil {
		log.Error(err, "failed to listen grpc address")
		return err
	}

	g.Go(func() error {
		log.Info("starting grpc endpoint", "address", grpcBindAddr, "runtime", conf.Runtime)
		if err := grpcServer.Serve(lis); err != nil {
			log.Error(err, "failed to start grpc endpoint")
			grpcServer.Stop()
			return err
		}
		return nil
	})

	return g.Wait()
}
