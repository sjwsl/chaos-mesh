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
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/mock"
)

var _ = Describe("ipset server", func() {
	defer mock.With("MockContainerdClient", &MockClient{})()
	c, _ := CreateContainerRuntimeInfoClient(containerRuntimeContainerd)
	s := &daemonServer{c}

	Context("createIPSet", func() {
		It("should work", func() {
			defer mock.With("MockWithNetNs", func(ctx context.Context, ns string, cmd string, args ...string) *exec.Cmd {
				Expect(ns).To(Equal("nsPath"))
				Expect(cmd).To(Equal("ipset"))
				Expect(args[0]).To(Equal("create"))
				Expect(args[1]).To(Equal("name"))
				Expect(args[2]).To(Equal("hash:net"))
				return exec.Command("echo", "mock command")
			})()
			err := s.createIPSet(context.TODO(), "nsPath", "name")
			Expect(err).To(BeNil())
		})

		It("should work since ipset exist", func() {
			// The mockfail.sh will fail only once
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
cat > /tmp/mockfail.sh << EOF
#! /bin/sh
exit 0
EOF
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockWithNetNs", func(context.Context, string, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", ipsetExistErr)
			})()
			err = s.createIPSet(context.TODO(), "nsPath", "name")
			Expect(err).To(BeNil())
		})

		It("shoud fail on the first command", func() {
			// The mockfail.sh will fail
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockWithNetNs", func(context.Context, string, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", "fail msg")
			})()
			err = s.createIPSet(context.TODO(), "nsPath", "name")
			Expect(err).ToNot(BeNil())
		})

		It("shoud fail on the second command", func() {
			// The mockfail.sh will fail
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockWithNetNs", func(context.Context, string, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", ipsetExistErr)
			})()
			err = s.createIPSet(context.TODO(), "nsPath", "name")
			Expect(err).ToNot(BeNil())
		})
	})

	Context("addIpsToIPSet", func() {
		It("should work", func() {
			defer mock.With("MockWithNetNs", func(context.Context, string, string, ...string) *exec.Cmd {
				return exec.Command("echo", "mock command")
			})()
			err := s.addCIDRsToIPSet(context.TODO(), "nsPath", "name", []string{"1.1.1.1"})
			Expect(err).To(BeNil())
		})

		It("should work since ip exist", func() {
			// The mockfail.sh will fail
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockWithNetNs", func(context.Context, string, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", ipExistErr)
			})()
			err = s.addCIDRsToIPSet(context.TODO(), "nsPath", "name", []string{"1.1.1.1"})
			Expect(err).To(BeNil())
		})

		It("should fail", func() {
			// The mockfail.sh will fail
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockWithNetNs", func(context.Context, string, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", "fail msg")
			})()
			err = s.addCIDRsToIPSet(context.TODO(), "nsPath", "name", []string{"1.1.1.1"})
			Expect(err).ToNot(BeNil())
		})
	})

	Context("renameIPSet", func() {
		It("should work", func() {
			defer mock.With("MockWithNetNs", func(context.Context, string, string, ...string) *exec.Cmd {
				return exec.Command("echo", "mock command")
			})()
			err := s.renameIPSet(context.TODO(), "nsPath", "name", "newname")
			Expect(err).To(BeNil())
		})

		It("should work since ipset exist", func() {
			// The mockfail.sh will fail only once
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
cat > /tmp/mockfail.sh << EOF
#! /bin/sh
exit 0
EOF
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockWithNetNs", func(context.Context, string, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", ipsetNewNameExistErr)
			})()
			err = s.renameIPSet(context.TODO(), "nsPath", "name", "newname")
			Expect(err).To(BeNil())
		})

		It("shoud fail on the first command", func() {
			// The mockfail.sh will fail
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockWithNetNs", func(context.Context, string, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", "fail msg")
			})()
			err = s.renameIPSet(context.TODO(), "nsPath", "name", "newname")
			Expect(err).ToNot(BeNil())
		})

		It("shoud fail on the second command", func() {
			// The mockfail.sh will fail
			err := ioutil.WriteFile("/tmp/mockfail.sh", []byte(`#! /bin/sh
echo $1
exit 1
			`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/mockfail.sh")
			defer mock.With("MockWithNetNs", func(context.Context, string, string, ...string) *exec.Cmd {
				return exec.Command("/tmp/mockfail.sh", ipsetExistErr)
			})()
			err = s.renameIPSet(context.TODO(), "nsPath", "name", "newname")
			Expect(err).ToNot(BeNil())
		})
	})

	Context("FlushIPSet", func() {
		It("should work", func() {
			defer mock.With("MockWithNetNs", func(context.Context, string, string, ...string) *exec.Cmd {
				return exec.Command("echo", "mock command")
			})()
			_, err := s.FlushIpSet(context.TODO(), &pb.IpSetRequest{
				Ipset: &pb.IpSet{
					Name:  "ipset-name",
					Cidrs: []string{"1.1.1.1/32"},
				},
				ContainerId: "containerd://container-id",
			})
			Expect(err).To(BeNil())
		})

		It("should fail on get pid", func() {
			const errorStr = "mock get pid error"
			defer mock.With("TaskError", errors.New(errorStr))()
			_, err := s.FlushIpSet(context.TODO(), &pb.IpSetRequest{
				Ipset: &pb.IpSet{
					Name:  "ipset-name",
					Cidrs: []string{"1.1.1.1/32"},
				},
				ContainerId: "containerd://container-id",
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(errorStr))
		})
	})
})
