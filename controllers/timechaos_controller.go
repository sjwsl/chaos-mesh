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

package controllers

import (
	"github.com/go-logr/logr"

	"github.com/pingcap/chaos-mesh/controllers/timechaos"

	v1alpha1 "github.com/pingcap/chaos-mesh/api/v1alpha1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TimeChaosReconciler reconciles a TimeChaos object
type TimeChaosReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=pingcap.com,resources=timechaos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pingcap.com,resources=timechaos/status,verbs=get;update;patch

// Reconcile reconciles a request from controller
func (r *TimeChaosReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("reconciler", "timechaos")

	reconciler := timechaos.Reconciler{
		Client: r.Client,
		Log:    logger,
	}

	return reconciler.Reconcile(req)
}

// SetupWithManager setups a time chaos reconciler on controller-manager
func (r *TimeChaosReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.TimeChaos{}).
		Complete(r)
}
