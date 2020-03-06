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

	"github.com/pingcap/chaos-mesh/controllers/kernelchaos"

	v1alpha1 "github.com/pingcap/chaos-mesh/api/v1alpha1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KernelChaosReconciler reconciles a KernelChaos object
type KernelChaosReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=pingcap.com,resources=kernelchaos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pingcap.com,resources=kernelchaos/status,verbs=get;update;patch

// Reconcile reconciles a request from controller
func (r *KernelChaosReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("reconciler", "kernelchaos")

	reconciler := kernelchaos.Reconciler{
		Client: r.Client,
		Log:    logger,
	}

	return reconciler.Reconcile(req)
}

// SetupWithManager sets up a kernel chaos reconciler on controller-manager
func (r *KernelChaosReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.KernelChaos{}).
		Complete(r)
}
