// Copyright 2019 Chaos Mesh Authors.
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
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// These tests are written in BDD-style using Ginkgo framework. Refer to
// http://onsi.github.io/ginkgo to learn more.

var _ = Describe("NetworkChaos", func() {
	var (
		key              types.NamespacedName
		created, fetched *NetworkChaos
	)

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Create API", func() {
		It("should create an object successfully", func() {
			key = types.NamespacedName{
				Name:      "foo",
				Namespace: "default",
			}

			created = &NetworkChaos{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Spec: NetworkChaosSpec{
					Mode:   OnePodMode,
					Action: DelayAction,
				},
			}

			By("creating an API obj")
			Expect(k8sClient.Create(context.TODO(), created)).To(Succeed())

			fetched = &NetworkChaos{}
			Expect(k8sClient.Get(context.TODO(), key, fetched)).To(Succeed())
			Expect(fetched).To(Equal(created))

			By("deleting the created object")
			Expect(k8sClient.Delete(context.TODO(), created)).To(Succeed())
			Expect(k8sClient.Get(context.TODO(), key, created)).ToNot(Succeed())
		})

		It("should set next start time successfully", func() {
			nwChaos := &NetworkChaos{}
			nTime := time.Now()
			nwChaos.SetNextStart(nTime)
			Expect(nwChaos.GetNextStart()).To(Equal(nTime))
		})

		It("should set recover time successfully", func() {
			nwChaos := &NetworkChaos{}
			nTime := time.Now()
			nwChaos.SetNextRecover(nTime)
			Expect(nwChaos.GetNextRecover()).To(Equal(nTime))
		})
	})

	Context("convertUnitToBytes", func() {
		It("should convert number with unit successfully", func() {
			n, err := convertUnitToBytes("  10   mbPs  ")
			Expect(err).Should(Succeed())
			Expect(n).To(Equal(uint64(10 * 1024 * 1024)))
		})

		It("should return error with invalid unit", func() {
			n, err := convertUnitToBytes(" 10 cpbs")
			Expect(err).Should(HaveOccurred())
			Expect(n).To(Equal(uint64(0)))
		})
	})
})
