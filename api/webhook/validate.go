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

package webhook

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/pingcap/chaos-mesh/pkg/webhook/validation"
)

//// +kubebuilder:webhook:path=/validate-v1alpha1-chaos,validating=true,failurePolicy=fail,groups="pingcap.com",resources=iochaos;podchaos;networkchaos;timechaos,verbs=create;update,versions=v1,name=chaos.validate

var validatelog = ctrl.Log.WithName("validate-webhook")

// ChaosValidator used to handle the validation admission request
type ChaosValidator struct{}

// Handle handles the requests from validation admission requests
func (v *ChaosValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	name := req.Name
	namespace := req.Namespace
	kind := req.Kind.Kind
	validatelog.Info(fmt.Sprintf("Receive validation req for obj[%s/%s/%s]", kind, namespace, name))
	return admission.Response{
		AdmissionResponse: *validation.ValidateChaos(&req.AdmissionRequest, kind),
	}
}
