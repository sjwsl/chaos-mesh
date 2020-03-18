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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// TimeChaos is the Schema for the timechaos API
type TimeChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a time chaos experiment
	Spec TimeChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the time chaos experiment
	Status TimeChaosStatus `json:"status"`
}

// TimeChaosSpec defines the desired state of TimeChaos
type TimeChaosSpec struct {
	// Mode defines the mode to run chaos action.
	// Supported mode: one / all / fixed / fixed-percent / random-max-percent
	Mode PodMode `json:"mode"`

	// Value is required when the mode is set to `FixedPodMode` / `FixedPercentPodMod` / `RandomMaxPercentPodMod`.
	// If `FixedPodMode`, provide an integer of pods to do chaos action.
	// If `FixedPercentPodMod`, provide a number from 0-100 to specify the max % of pods the server can do chaos action.
	// If `RandomMaxPercentPodMod`,  provide a number from 0-100 to specify the % of pods to do chaos action
	// +optional
	Value string `json:"value"`

	// Selector is used to select pods that are used to inject chaos action.
	Selector SelectorSpec `json:"selector"`

	// TimeOffset defines the delta time of injected program
	TimeOffset TimeOffset `json:"timeOffset"`

	// ClockIds defines all affected clock id
	// All available options are ["CLOCK_REALTIME","CLOCK_MONOTONIC","CLOCK_PROCESS_CPUTIME_ID","CLOCK_THREAD_CPUTIME_ID",
	// "CLOCK_MONOTONIC_RAW","CLOCK_REALTIME_COARSE","CLOCK_MONOTONIC_COARSE","CLOCK_BOOTTIME","CLOCK_REALTIME_ALARM",
	// "CLOCK_BOOTTIME_ALARM"]
	// Default value is ["CLOCK_REALTIME"]
	ClockIds []string `json:"clockIds,omitempty"`

	// ContainerName indicates the name of affected container.
	// If not set, all containers will be injected
	// +optional
	ContainerNames []string `json:"containerNames,omitempty"`

	// Duration represents the duration of the chaos action
	Duration *string `json:"duration,omitempty"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`

	// Next time when this action will be applied again
	// +optional
	NextStart *metav1.Time `json:"nextStart,omitempty"`

	// Next time when this action will be recovered
	// +optional
	NextRecover *metav1.Time `json:"nextRecover,omitempty"`
}

// SetDefaultValue will set default value for empty fields
func (in *TimeChaos) SetDefaultValue() {
	if in.Spec.ClockIds == nil || len(in.Spec.ClockIds) == 0 {
		in.Spec.ClockIds = []string{"CLOCK_REALTIME"}
	}
}

// GetSelector is a getter for Selector (for implementing SelectSpec)
func (in *TimeChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

// GetMode is a getter for Mode (for implementing SelectSpec)
func (in *TimeChaosSpec) GetMode() PodMode {
	return in.Mode
}

// GetValue is a getter for Value (for implementing SelectSpec)
func (in *TimeChaosSpec) GetValue() string {
	return in.Value
}

// TimeOffset defines the delta time of injected program
// As `clock_gettime` return a struct contains two field: `tv_sec` and `tv_nsec`.
// `Sec` is the offset of seconds, corresponding to `tv_sec` field.
// `NSec` is the offset of nanoseconds, corresponding to `tv_nsec` field.
type TimeOffset struct {
	Sec  int64 `json:"sec"`
	NSec int64 `json:"nsec"`
}

// TimeChaosStatus defines the observed state of TimeChaos
type TimeChaosStatus struct {
	ChaosStatus `json:",inline"`
}

// GetDuration gets the duration of TimeChaos
func (in *TimeChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetNextStart gets NextStart field of TimeChaos
func (in *TimeChaos) GetNextStart() time.Time {
	if in.Spec.NextStart == nil {
		return time.Time{}
	}
	return in.Spec.NextStart.Time
}

// SetNextStart sets NextStart field of TimeChaos
func (in *TimeChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Spec.NextStart = nil
		return
	}

	if in.Spec.NextStart == nil {
		in.Spec.NextStart = &metav1.Time{}
	}
	in.Spec.NextStart.Time = t
}

// GetNextRecover get NextRecover field of TimeChaos
func (in *TimeChaos) GetNextRecover() time.Time {
	if in.Spec.NextRecover == nil {
		return time.Time{}
	}
	return in.Spec.NextRecover.Time
}

// SetNextRecover sets NextRecover field of TimeChaos
func (in *TimeChaos) SetNextRecover(t time.Time) {
	if t.IsZero() {
		in.Spec.NextRecover = nil
		return
	}

	if in.Spec.NextRecover == nil {
		in.Spec.NextRecover = &metav1.Time{}
	}
	in.Spec.NextRecover.Time = t
}

// GetScheduler returns the scheduler of TimeChaos
func (in *TimeChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetStatus returns the status of TimeChaos
func (in *TimeChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// IsDeleted returns whether this resource has been deleted
func (in *TimeChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// +kubebuilder:object:root=true

// TimeChaosList contains a list of TimeChaos
type TimeChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TimeChaos `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TimeChaos{}, &TimeChaosList{})
}
