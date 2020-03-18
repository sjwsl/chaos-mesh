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

package v1alpha1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var kernelchaoslog = logf.Log.WithName("kernelchaos-resource")

// SetupWebhookWithManager setup KernelChaos's webhook with manager
func (in *KernelChaos) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-pingcap-com-v1alpha1-kernelchaos,mutating=true,failurePolicy=fail,groups=pingcap.com,resources=kernelchaos,verbs=create;update,versions=v1alpha1,name=mkernelchaos.kb.io

var _ webhook.Defaulter = &KernelChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *KernelChaos) Default() {
	kernelchaoslog.Info("default", "name", in.Name)

	in.Spec.Selector.DefaultNamespace(in.GetNamespace())
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-pingcap-com-v1alpha1-kernelchaos,mutating=false,failurePolicy=fail,groups=pingcap.com,resources=kernelchaos,versions=v1alpha1,name=vkernelchaos.kb.io

var _ ChaosValidator = &KernelChaos{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *KernelChaos) ValidateCreate() error {
	kernelchaoslog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *KernelChaos) ValidateUpdate(old runtime.Object) error {
	kernelchaoslog.Info("validate update", "name", in.Name)
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *KernelChaos) ValidateDelete() error {
	kernelchaoslog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

// Validate validates chaos object
func (in *KernelChaos) Validate() error {
	specField := field.NewPath("spec")
	errLst := in.ValidateScheduler(specField)

	if len(errLst) > 0 {
		return fmt.Errorf(errLst.ToAggregate().Error())
	}
	return nil
}

// ValidateScheduler validates the scheduler and duration
func (in *KernelChaos) ValidateScheduler(root *field.Path) field.ErrorList {
	if in.Spec.Duration != nil && in.Spec.Scheduler != nil {
		return nil
	} else if in.Spec.Duration == nil && in.Spec.Scheduler == nil {
		return nil
	}

	allErrs := field.ErrorList{}
	schedulerField := root.Child("scheduler")

	allErrs = append(allErrs, field.Invalid(schedulerField, in.Spec.Scheduler, ValidateSchedulerError))
	return allErrs
}
