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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pingcap/chaos-mesh/pkg/mock"
)

var _ = Describe("chaosdaemon util", func() {
	Context("DockerClient GetPidFromContainerID", func() {
		It("should return the magic number 9527", func() {
			defer mock.With("pid", int(9527))()
			m := &MockClient{}
			c := DockerClient{client: m}
			pid, err := c.GetPidFromContainerID(context.TODO(), "docker://valid-container-id")
			Expect(err).To(BeNil())
			Expect(pid).To(Equal(uint32(9527)))
		})

		It("should error with wrong protocol", func() {
			m := &MockClient{}
			c := DockerClient{client: m}
			_, err := c.GetPidFromContainerID(context.TODO(), "containerd://this-is-a-wrong-protocol")
			Expect(err).NotTo(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring(fmt.Sprintf("expected %s but got", dockerProtocolPrefix)))
		})

		It("should error on ContainerInspectError", func() {
			errorStr := "this is a mocked error"
			defer mock.With("ContainerInspectError", errors.New(errorStr))()
			m := &MockClient{}
			c := DockerClient{client: m}
			_, err := c.GetPidFromContainerID(context.TODO(), "docker://valid-container-id")
			Expect(err).NotTo(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(Equal(errorStr))
		})
	})

	Context("ContainerdClient GetPidFromContainerID", func() {
		It("should return the magic number 9527", func() {
			defer mock.With("pid", int(9527))()
			m := &MockClient{}
			c := ContainerdClient{client: m}
			pid, err := c.GetPidFromContainerID(context.TODO(), "containerd://valid-container-id")
			Expect(err).To(BeNil())
			Expect(pid).To(Equal(uint32(9527)))
		})

		It("should error with wrong protocol", func() {
			m := &MockClient{}
			c := ContainerdClient{client: m}
			_, err := c.GetPidFromContainerID(context.TODO(), "docker://this-is-a-wrong-protocol")
			Expect(err).NotTo(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring(fmt.Sprintf("expected %s but got", containerdProtocolPrefix)))
		})

		It("should error with specified string", func() {
			errorStr := "this is a mocked error"
			mock.With("LoadContainerError", errors.New(errorStr))
			m := &MockClient{}
			c := ContainerdClient{client: m}
			_, err := c.GetPidFromContainerID(context.TODO(), "containerd://valid-container-id")
			Expect(err).NotTo(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(Equal(errorStr))
			mock.Reset("LoadContainerError")

			mock.With("TaskError", errors.New(errorStr))
			m = &MockClient{}
			c = ContainerdClient{client: m}
			_, err = c.GetPidFromContainerID(context.TODO(), "containerd://valid-container-id")
			Expect(err).NotTo(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(Equal(errorStr))
			mock.Reset("TaskError")
		})
	})

	Context("CreateContainerRuntimeInfoClient", func() {
		It("should work", func() {
			_, err := CreateContainerRuntimeInfoClient(containerRuntimeDocker)
			Expect(err).To(BeNil())

			defer mock.With("MockContainerdClient", &MockClient{})()
			_, err = CreateContainerRuntimeInfoClient(containerRuntimeContainerd)
			Expect(err).To(BeNil())
		})

		It("should error on newContaineredClient", func() {
			errorStr := "this is a mocked error"
			defer mock.With("NewContainerdClientError", errors.New(errorStr))()
			_, err := CreateContainerRuntimeInfoClient(containerRuntimeContainerd)
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(Equal(errorStr))
		})
	})

	Context("DockerClient ContainerKillByContainerID", func() {
		It("should work", func() {
			m := &MockClient{}
			c := DockerClient{client: m}
			err := c.ContainerKillByContainerID(context.TODO(), "docker://valid-container-id")
			Expect(err).To(BeNil())
		})

		It("should error on ContainerKill", func() {
			errorStr := "this is a mocked error on ContainerKill"
			m := &MockClient{}
			c := DockerClient{client: m}
			defer mock.With("ContainerKillError", errors.New(errorStr))()
			err := c.ContainerKillByContainerID(context.TODO(), "docker://valid-container-id")
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(Equal(errorStr))
		})

		It("should error on wrong protocol", func() {
			m := &MockClient{}
			c := DockerClient{client: m}
			err := c.ContainerKillByContainerID(context.TODO(), "containerd://this-is-a-wrong-protocol")
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring(fmt.Sprintf("expected %s but got", dockerProtocolPrefix)))
		})

		It("should error on short protocol", func() {
			m := &MockClient{}
			c := DockerClient{client: m}
			err := c.ContainerKillByContainerID(context.TODO(), "dock:")
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("is not a docker container id"))
		})
	})

	Context("ContainerdClient ContainerKillByContainerID", func() {
		It("should work", func() {
			m := &MockClient{}
			c := ContainerdClient{client: m}
			err := c.ContainerKillByContainerID(context.TODO(), "containerd://valid-container-id")
			Expect(err).To(BeNil())
		})

		errorPoints := []string{"LoadContainer", "Task", "Kill"}
		for _, e := range errorPoints {
			It(fmt.Sprintf("should error on %s", e), func() {
				errorStr := fmt.Sprintf("this is a mocked error on %s", e)
				m := &MockClient{}
				c := ContainerdClient{client: m}
				defer mock.With(e+"Error", errors.New(errorStr))()
				err := c.ContainerKillByContainerID(context.TODO(), "containerd://valid-container-id")
				Expect(err).ToNot(BeNil())
				Expect(fmt.Sprintf("%s", err)).To(Equal(errorStr))
			})
		}

		It("should error on wrong protocol", func() {
			m := &MockClient{}
			c := ContainerdClient{client: m}
			err := c.ContainerKillByContainerID(context.TODO(), "docker://this-is-a-wrong-protocol")
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring(fmt.Sprintf("expected %s but got", containerdProtocolPrefix)))
		})

		It("should error on short protocol", func() {
			m := &MockClient{}
			c := ContainerdClient{client: m}
			err := c.ContainerKillByContainerID(context.TODO(), "dock:")
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("is not a containerd container id"))
		})
	})
})
