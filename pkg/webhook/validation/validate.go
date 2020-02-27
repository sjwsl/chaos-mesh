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

package validation

import (
	"encoding/json"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
)

const (
	invalidConfigurationMsg = "invalid configuration"
)

var log = ctrl.Log.WithName("validate-webhook")

// ValidateChaos handles the validation for chaos api
func ValidateChaos(res *admissionv1beta1.AdmissionRequest, kind string) *admissionv1beta1.AdmissionResponse {
	var err error
	var permitted bool
	var msg string
	switch kind {
	case "PodChaos":
		permitted, msg, err = validatePodchaos(res.Object.Raw)
	case "NetworkChaos":
		permitted, msg, err = validateNetworkChaos(res.Object.Raw)
	case "IoChaos":
		permitted, msg, err = validateIoChaos(res.Object.Raw)
	case "TimeChaos":
		permitted, msg, err = validateTimeChaos(res.Object.Raw)
	default:
		log.Error(err, "Could not unmarshal raw object")
	}
	if err != nil {
		return &admissionv1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	return &admissionv1beta1.AdmissionResponse{
		Allowed: permitted,
		Result: &metav1.Status{
			Message: msg,
		},
	}
}

func validatePodchaos(rawObj []byte) (bool, string, error) {
	var chaos v1alpha1.PodChaos
	if err := json.Unmarshal(rawObj, &chaos); err != nil {
		return false, "", err
	}
	return chaos.Validate()
}

func validateIoChaos(rawObj []byte) (bool, string, error) {
	var chaos v1alpha1.IoChaos
	if err := json.Unmarshal(rawObj, &chaos); err != nil {
		return false, "", err
	}
	return chaos.Validate()
}

func validateNetworkChaos(rawObj []byte) (bool, string, error) {
	var chaos v1alpha1.NetworkChaos
	if err := json.Unmarshal(rawObj, &chaos); err != nil {
		return false, "", err
	}
	return chaos.Validate()
}

func validateTimeChaos(rawObj []byte) (bool, string, error) {
	var chaos v1alpha1.TimeChaos
	if err := json.Unmarshal(rawObj, &chaos); err != nil {
		return false, "", err
	}
	return chaos.Validate()
}
