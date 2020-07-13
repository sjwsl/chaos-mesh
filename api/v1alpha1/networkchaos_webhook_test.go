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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("networkchaos_webhook", func() {
	Context("Defaulter", func() {
		It("set default namespace selector", func() {
			networkchaos := &NetworkChaos{
				ObjectMeta: metav1.ObjectMeta{Namespace: metav1.NamespaceDefault},
			}
			networkchaos.Default()
			Expect(networkchaos.Spec.Selector.Namespaces[0]).To(Equal(metav1.NamespaceDefault))
		})

		It("set default DelaySpec", func() {
			networkchaos := &NetworkChaos{
				ObjectMeta: metav1.ObjectMeta{Namespace: metav1.NamespaceDefault},
				Spec: NetworkChaosSpec{
					Delay: &DelaySpec{
						Latency: "90ms",
					},
				},
			}
			networkchaos.Default()
			Expect(networkchaos.Spec.Delay.Correlation).To(Equal(DefaultCorrelation))
			Expect(networkchaos.Spec.Delay.Jitter).To(Equal(DefaultJitter))
		})
	})
	Context("ChaosValidator of networkchaos", func() {
		It("Validate", func() {

			type TestCase struct {
				name    string
				chaos   NetworkChaos
				execute func(chaos *NetworkChaos) error
				expect  string
			}
			duration := "400s"
			tcs := []TestCase{
				{
					name: "simple ValidateCreate",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo1",
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "",
				},
				{
					name: "simple ValidateUpdate",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo2",
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateUpdate(chaos)
					},
					expect: "",
				},
				{
					name: "simple ValidateDelete",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo3",
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateDelete()
					},
					expect: "",
				},
				{
					name: "only define the Scheduler",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo4",
						},
						Spec: NetworkChaosSpec{
							Scheduler: &SchedulerSpec{
								Cron: "@every 10m",
							},
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "only define the Duration",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo5",
						},
						Spec: NetworkChaosSpec{
							Duration: &duration,
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the delay",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo6",
						},
						Spec: NetworkChaosSpec{
							Delay: &DelaySpec{
								Latency:     "1S",
								Jitter:      "1S",
								Correlation: "num",
							},
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the reorder",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo7",
						},
						Spec: NetworkChaosSpec{
							Delay: &DelaySpec{
								Reorder: &ReorderSpec{
									Reorder:     "num",
									Correlation: "num",
								},
							},
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the loss",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo8",
						},
						Spec: NetworkChaosSpec{
							Loss: &LossSpec{
								Loss:        "num",
								Correlation: "num",
							},
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the duplicate",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo9",
						},
						Spec: NetworkChaosSpec{
							Duplicate: &DuplicateSpec{
								Duplicate:   "num",
								Correlation: "num",
							},
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the corrupt",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo10",
						},
						Spec: NetworkChaosSpec{
							Corrupt: &CorruptSpec{
								Corrupt:     "num",
								Correlation: "num",
							},
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the bandwidth",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo11",
						},
						Spec: NetworkChaosSpec{
							Bandwidth: &BandwidthSpec{
								Rate: "10",
							},
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate the target",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo12",
						},
						Spec: NetworkChaosSpec{
							Target: &Target{
								TargetMode:  FixedPodMode,
								TargetValue: "0",
							},
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate direction and externalTargets",
					chaos: NetworkChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo12",
						},
						Spec: NetworkChaosSpec{
							Direction:       From,
							ExternalTargets: []string{"8.8.8.8"},
						},
					},
					execute: func(chaos *NetworkChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
			}

			for _, tc := range tcs {
				err := tc.execute(&tc.chaos)
				if tc.expect == "error" {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			}
		})
	})
})
