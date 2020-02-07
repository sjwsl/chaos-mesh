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

package reconciler

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/pingcap/chaos-mesh/pkg/apiinterface"
)

// InnerObject is basic Object for the Reconciler
type InnerObject interface {
	IsDeleted() bool
	apiinterface.StatefulObject
}

// InnerReconciler is interface for reconciler
type InnerReconciler interface {

	// Apply means the reconciler perform the chaos action
	Apply(ctx context.Context, req ctrl.Request, chaos InnerObject) error

	// Recover means the reonciler recover the chaos action
	Recover(ctx context.Context, req ctrl.Request, chaos InnerObject) error

	// Object would return the instance of chaos
	Object() InnerObject
}
