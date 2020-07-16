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

package chaosdaemon

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var _ = Describe("netem server", func() {
	defer mock.With("MockContainerdClient", &MockClient{})()
	c, _ := CreateContainerRuntimeInfoClient(containerRuntimeContainerd)
	s := &daemonServer{c}

	Context("SetTbf", func() {
		It("should work", func() {
			const ignore = true
			defer mock.With("TbfApplyError", ignore)()
			_, err := s.SetTbf(context.TODO(), &pb.TbfRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).To(BeNil())
		})

		It("should fail on get pid", func() {
			const errorStr = "mock error on Task()"
			defer mock.With("TaskError", errors.New(errorStr))()
			_, err := s.SetTbf(context.TODO(), &pb.TbfRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring(errorStr))
		})

		It("should fail on applyNetem", func() {
			const errorStr = "mock error on applyNetem()"
			defer mock.With("TbfApplyError", errors.New(errorStr))()
			_, err := s.SetTbf(context.TODO(), &pb.TbfRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring(errorStr))
		})
	})

	Context("DeleteTbf", func() {
		It("should work", func() {
			const ignore = true
			defer mock.With("TbfDeleteError", ignore)()
			_, err := s.DeleteTbf(context.TODO(), &pb.TbfRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).To(BeNil())
		})

		It("should fail on get pid", func() {
			const errorStr = "mock error on Task()"
			defer mock.With("TaskError", errors.New(errorStr))()
			_, err := s.DeleteTbf(context.TODO(), &pb.TbfRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring(errorStr))
		})

		It("should fail on applyNetem", func() {
			const errorStr = "mock error on applyNetem()"
			defer mock.With("TbfDeleteError", errors.New(errorStr))()
			_, err := s.DeleteTbf(context.TODO(), &pb.TbfRequest{
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring(errorStr))
		})
	})
})
