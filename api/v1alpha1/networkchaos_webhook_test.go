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
			Expect(networkchaos.Spec.Target.TargetSelector.Namespaces[0]).To(Equal(metav1.NamespaceDefault))
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
