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

package podfailure

import (
	"context"

	"github.com/go-logr/logr"

	"github.com/pingcap/chaos-mesh/controllers/common"
	"github.com/pingcap/chaos-mesh/controllers/podchaos/podfailure"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CommonReconciler is reconciler for podfailure
type CommonReconciler struct {
	*podfailure.Reconciler
}

// NewCommonReconciler would create reconciler for common chaos
func NewCommonReconciler(c client.Client, log logr.Logger, req ctrl.Request) *CommonReconciler {
	r := &podfailure.Reconciler{
		Client: c,
		Log:    log,
	}
	return &CommonReconciler{
		Reconciler: r,
	}
}

// Apply would perform common chaos for podchaos
func (r *CommonReconciler) Apply(ctx context.Context, req ctrl.Request, chaos common.InnerCommonObject) error {
	return r.Perform(ctx, req, chaos)
}

// Recover would recover the common chaos for podchaos
func (r *CommonReconciler) Recover(ctx context.Context, req ctrl.Request, chaos common.InnerCommonObject) error {
	return r.Clean(ctx, req, chaos)
}

// Object implement common.Object
func (r *CommonReconciler) Object() common.InnerCommonObject {
	return r.Instance()
}
