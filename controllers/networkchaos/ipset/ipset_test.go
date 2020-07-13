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

package ipset

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cmv1alpha1 "github.com/pingcap/chaos-mesh/api/v1alpha1"
)

func Test_generateIpSetName(t *testing.T) {
	g := NewWithT(t)
	postfix := "alongpostfix"

	t.Run("name with postfix", func(t *testing.T) {
		chaosName := "test"

		networkChaos := &cmv1alpha1.NetworkChaos{
			ObjectMeta: metav1.ObjectMeta{
				Name: chaosName,
			},
		}

		name := GenerateIPSetName(networkChaos, postfix)

		g.Expect(name).Should(Equal(chaosName + "_" + postfix))
	})

	t.Run("length equal 27", func(t *testing.T) {
		networkChaos := &cmv1alpha1.NetworkChaos{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-metav1object",
			},
		}

		name := GenerateIPSetName(networkChaos, postfix)

		g.Expect(len(name)).Should(Equal(27))
	})
}
