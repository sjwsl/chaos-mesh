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
	"time"

	"github.com/docker/go-units"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Stress chaos is a chaos to generate plenty of stresses over a collection of pods.

// KindStressChaos is the kind for stress chaos
const KindStressChaos = "StressChaos"

func init() {
	all.register(KindStressChaos, &ChaosKind{
		Chaos:     &StressChaos{},
		ChaosList: &StressChaosList{},
	})
}

// +kubebuilder:object:root=true

// StressChaos is the Schema for the stresschaos API
type StressChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a time chaos experiment
	Spec StressChaosSpec `json:"spec"`

	// +optional
	// Most recently observed status of the time chaos experiment
	Status StressChaosStatus `json:"status"`
}

// StressChaosSpec defines the desired state of StressChaos
type StressChaosSpec struct {
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

	// Stressors defines plenty of stressors supported to stress system components out.
	// You can use one or more of them to make up various kinds of stresses. At least
	// one of the stressors should be specified.
	// +optional
	Stressors *Stressors `json:"stressors,omitempty"`

	// StressngStressors defines plenty of stressors just like `Stressors` except that it's an experimental
	// feature and more powerful. You can define stressors in `stress-ng` (see also `man stress-ng`) dialect,
	// however not all of the supported stressors are well tested. It maybe retired in later releases. You
	// should always use `Stressors` to define the stressors and use this only when you want more stressors
	// unsupported by `Stressors`. When both `StressngStressors` and `Stressors` are defined, `StressngStressors`
	// wins.
	// +optional
	StressngStressors string `json:"stressngStressors,omitempty"`

	// Duration represents the duration of the chaos action
	// +optional
	Duration *string `json:"duration,omitempty"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	// +optional
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`
}

// GetSelector is a getter for Selector (for implementing SelectSpec)
func (in *StressChaosSpec) GetSelector() SelectorSpec {
	return in.Selector
}

// GetMode is a getter for Mode (for implementing SelectSpec)
func (in *StressChaosSpec) GetMode() PodMode {
	return in.Mode
}

// GetValue is a getter for Value (for implementing SelectSpec)
func (in *StressChaosSpec) GetValue() string {
	return in.Value
}

// StressChaosStatus defines the observed state of StressChaos
type StressChaosStatus struct {
	ChaosStatus `json:",inline"`
	// Instances always specifies stressing instances
	// +optional
	Instances map[string]StressInstance `json:"instances,omitempty"`
}

// StressInstance is an instance generates stresses
type StressInstance struct {
	// UID is the instance identifier
	// +optional
	UID string `json:"uid"`
	// StartTime specifies when the instance starts
	// +optional
	StartTime *metav1.Time `json:"startTime"`
}

// GetDuration gets the duration of StressChaos
func (in *StressChaos) GetDuration() (*time.Duration, error) {
	if in.Spec.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Spec.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

// GetNextStart gets NextStart field of StressChaos
func (in *StressChaos) GetNextStart() time.Time {
	if in.Status.Scheduler.NextStart == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextStart.Time
}

// SetNextStart sets NextStart field of StressChaos
func (in *StressChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextStart = nil
		return
	}

	if in.Status.Scheduler.NextStart == nil {
		in.Status.Scheduler.NextStart = &metav1.Time{}
	}
	in.Status.Scheduler.NextStart.Time = t
}

// GetNextRecover get NextRecover field of StressChaos
func (in *StressChaos) GetNextRecover() time.Time {
	if in.Status.Scheduler.NextRecover == nil {
		return time.Time{}
	}
	return in.Status.Scheduler.NextRecover.Time
}

// SetNextRecover sets NextRecover field of StressChaos
func (in *StressChaos) SetNextRecover(t time.Time) {
	if t.IsZero() {
		in.Status.Scheduler.NextRecover = nil
		return
	}

	if in.Status.Scheduler.NextRecover == nil {
		in.Status.Scheduler.NextRecover = &metav1.Time{}
	}
	in.Status.Scheduler.NextRecover.Time = t
}

// GetScheduler returns the scheduler of StressChaos
func (in *StressChaos) GetScheduler() *SchedulerSpec {
	return in.Spec.Scheduler
}

// GetStatus returns the status of StressChaos
func (in *StressChaos) GetStatus() *ChaosStatus {
	return &in.Status.ChaosStatus
}

// IsDeleted returns whether this resource has been deleted
func (in *StressChaos) IsDeleted() bool {
	return !in.DeletionTimestamp.IsZero()
}

// IsPaused returns whether this resource has been paused
func (in *StressChaos) IsPaused() bool {
	if in.Annotations == nil || in.Annotations[PauseAnnotationKey] != "true" {
		return false
	}
	return true
}

// GetChaos returns a chaos instance
func (in *StressChaos) GetChaos() *ChaosInstance {
	instance := &ChaosInstance{
		Name:      in.Name,
		Namespace: in.Namespace,
		Kind:      KindStressChaos,
		StartTime: in.CreationTimestamp.Time,
		Action:    "",
		Status:    string(in.GetStatus().Experiment.Phase),
	}
	if in.Spec.Duration != nil {
		instance.Duration = *in.Spec.Duration
	}
	if in.DeletionTimestamp != nil {
		instance.EndTime = in.DeletionTimestamp.Time
	}
	return instance
}

// Stressors defines plenty of stressors supported to stress system components out.
// You can use one or more of them to make up various kinds of stresses
type Stressors struct {
	// MemoryStressor stresses virtual memory out
	// +optional
	MemoryStressor *MemoryStressor `json:"memory,omitempty"`
	// CPUStressor stresses CPU out
	// +optional
	CPUStressor *CPUStressor `json:"cpu,omitempty"`
}

// Normalize the stressors to comply with stress-ng
func (in *Stressors) Normalize() (string, error) {
	stressors := ""
	if in.MemoryStressor != nil {
		stressors += fmt.Sprintf(" --vm %d --vm-keep", in.MemoryStressor.Workers)
		if len(in.MemoryStressor.Size) != 0 {
			if in.MemoryStressor.Size[len(in.MemoryStressor.Size)-1] != '%' {
				size, err := units.FromHumanSize(in.MemoryStressor.Size)
				if err != nil {
					return "", err
				}
				stressors += fmt.Sprintf(" --vm-bytes %d", size)
			} else {
				stressors += fmt.Sprintf("--vm-bytes %s",
					in.MemoryStressor.Size)
			}
		}

		if in.MemoryStressor.Options != nil {
			for _, v := range in.MemoryStressor.Options {
				stressors += fmt.Sprintf(" %v ", v)
			}
		}
	}
	if in.CPUStressor != nil {
		stressors += fmt.Sprintf(" --cpu %d", in.CPUStressor.Workers)
		if in.CPUStressor.Load != nil {
			stressors += fmt.Sprintf(" --cpu-load %d",
				*in.CPUStressor.Load)
		}

		if in.CPUStressor.Options != nil {
			for _, v := range in.CPUStressor.Options {
				stressors += fmt.Sprintf(" %v ", v)
			}
		}
	}
	return stressors, nil
}

// Stressor defines common configurations of a stressor
type Stressor struct {
	// Workers specifies N workers to apply the stressor.
	Workers int `json:"workers"`
}

// MemoryStressor defines how to stress memory out
type MemoryStressor struct {
	Stressor `json:",inline"`

	// Size specifies N bytes consumed per vm worker, default is the total available memory.
	// One can specify the size as % of total available memory or in units of B, KB/KiB,
	// MB/MiB, GB/GiB, TB/TiB.
	// +optional
	Size string `json:"size,omitempty"`

	// extend stress-ng options
	// +optional
	Options []string `json:"options,omitempty"`
}

// CPUStressor defines how to stress CPU out
type CPUStressor struct {
	Stressor `json:",inline"`
	// Load specifies P percent loading per CPU worker. 0 is effectively a sleep (no load) and 100
	// is full loading.
	// +optional
	Load *int `json:"load,omitempty"`

	// extend stress-ng options
	// +optional
	Options []string `json:"options,omitempty"`
}

// +kubebuilder:object:root=true

// StressChaosList contains a list of StressChaos
type StressChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StressChaos `json:"items"`
}

// ListChaos returns a list of stress chaos
func (in *StressChaosList) ListChaos() []*ChaosInstance {
	res := make([]*ChaosInstance, 0, len(in.Items))
	for _, item := range in.Items {
		res = append(res, item.GetChaos())
	}
	return res
}

func init() {
	SchemeBuilder.Register(&StressChaos{}, &StressChaosList{})
}
