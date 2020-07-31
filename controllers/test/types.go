// Copyright 2020 Chaos Mesh Authors.
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

package test

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	chaosdaemon "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

// Assert *MockChaosDaemonClient implements chaosdaemon.ChaosDaemonClientInterface.
var _ utils.ChaosDaemonClientInterface = (*MockChaosDaemonClient)(nil)

// MockChaosDaemonClient implements ChaosDaemonClientInterface for unit testing
type MockChaosDaemonClient struct{}

// ExecStressors mocks executing pod stressors on chaos-daemon
func (c *MockChaosDaemonClient) ExecStressors(ctx context.Context, in *chaosdaemon.ExecStressRequest, opts ...grpc.CallOption) (*chaosdaemon.ExecStressResponse, error) {
	return nil, mockError("ExecStressors")
}

// CancelStressors mocks canceling pod stressors on chaos-daemon
func (c *MockChaosDaemonClient) CancelStressors(ctx context.Context, in *chaosdaemon.CancelStressRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("CancelStressors")
}

func (c *MockChaosDaemonClient) ContainerGetPid(ctx context.Context, in *chaosdaemon.ContainerRequest, opts ...grpc.CallOption) (*chaosdaemon.ContainerResponse, error) {
	if resp := mock.On("MockContainerGetPidResponse"); resp != nil {
		return resp.(*chaosdaemon.ContainerResponse), nil
	}
	return nil, mockError("ContainerGetPid")
}

func mockError(name string) error {
	if err := mock.On(fmt.Sprintf("Mock%sError", name)); err != nil {
		return err.(error)
	}
	return nil
}

func (c *MockChaosDaemonClient) SetNetem(ctx context.Context, in *chaosdaemon.NetemRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("SetNetem")
}

func (c *MockChaosDaemonClient) DeleteNetem(ctx context.Context, in *chaosdaemon.NetemRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("DeleteNetem")
}

func (c *MockChaosDaemonClient) SetTbf(ctx context.Context, in *chaosdaemon.TbfRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("SetTbf")
}

func (c *MockChaosDaemonClient) DeleteTbf(ctx context.Context, in *chaosdaemon.TbfRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("DeleteTbf")
}

func (c *MockChaosDaemonClient) AddQdisc(ctx context.Context, in *chaosdaemon.QdiscRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("AddQdisc")
}

func (c *MockChaosDaemonClient) DelQdisc(ctx context.Context, in *chaosdaemon.QdiscRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("DelQdisc")
}

func (c *MockChaosDaemonClient) AddEmatchFilter(ctx context.Context, in *chaosdaemon.EmatchFilterRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("AddEmatchFilter")
}

func (c *MockChaosDaemonClient) DelTcFilter(ctx context.Context, in *chaosdaemon.TcFilterRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("DelTcFilter")
}

func (c *MockChaosDaemonClient) FlushIPSets(ctx context.Context, in *chaosdaemon.IPSetsRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("FlushIPSets")
}

func (c *MockChaosDaemonClient) SetIptablesChains(ctx context.Context, in *chaosdaemon.IptablesChainsRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("SetIptablesChains")
}

func (c *MockChaosDaemonClient) SetTimeOffset(ctx context.Context, in *chaosdaemon.TimeRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("SetTimeOffset")
}

func (c *MockChaosDaemonClient) RecoverTimeOffset(ctx context.Context, in *chaosdaemon.TimeRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("RecoverTimeOffset")
}

func (c *MockChaosDaemonClient) ContainerKill(ctx context.Context, in *chaosdaemon.ContainerRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, mockError("ContainerKill")
}

func (c *MockChaosDaemonClient) Close() error {
	return mockError("CloseChaosDaemonClient")
}

func newPod(
	name string,
	status v1.PodPhase,
	namespace string,
	ans map[string]string,
	ls map[string]string,
	containerStatus v1.ContainerStatus,
) v1.Pod {
	return v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      ls,
			Annotations: ans,
		},
		Status: v1.PodStatus{
			Phase:             status,
			ContainerStatuses: []v1.ContainerStatus{containerStatus},
		},
		Spec: v1.PodSpec{
			InitContainers: []v1.Container{{Name: "fake-name", Image: "fake-image"}},
			Containers:     []v1.Container{{Name: "fake-name", Image: "fake-image"}},
		},
	}
}

// GenerateNPods is only for unit testing
func GenerateNPods(
	namePrefix string,
	n int,
	status v1.PodPhase,
	ns string,
	ans map[string]string,
	ls map[string]string,
	containerStatus v1.ContainerStatus,
) ([]runtime.Object, []v1.Pod) {
	var podObjects []runtime.Object
	var pods []v1.Pod
	for i := 0; i < n; i++ {
		pod := newPod(fmt.Sprintf("%s%d", namePrefix, i), status, ns, ans, ls, containerStatus)
		podObjects = append(podObjects, &pod)
		pods = append(pods, pod)
	}

	return podObjects, pods
}
