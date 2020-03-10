package utils

import (
	"context"

	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	chaosdaemon "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/mock"
)

// ChaosDaemonClientInterface represents the ChaosDaemonClient, it's used to simply unit test
type ChaosDaemonClientInterface interface {
	chaosdaemon.ChaosDaemonClient
	Close() error
}

// GrpcChaosDaemonClient would act like chaosdaemon.ChaosDaemonClient with a Close method
type GrpcChaosDaemonClient struct {
	chaosdaemon.ChaosDaemonClient
	conn *grpc.ClientConn
}

func (c *GrpcChaosDaemonClient) Close() error {
	return c.conn.Close()
}

// NewChaosDaemonClient would create ChaosDaemonClient
func NewChaosDaemonClient(ctx context.Context, c client.Client, pod *v1.Pod, port string) (ChaosDaemonClientInterface, error) {
	if cli := mock.On("MockChaosDaemonClient"); cli != nil {
		return cli.(ChaosDaemonClientInterface), nil
	}
	if err := mock.On("NewChaosDaemonClientError"); err != nil {
		return nil, err.(error)
	}

	cc, err := CreateGrpcConnection(ctx, c, pod, port)
	if err != nil {
		return nil, err
	}
	return &GrpcChaosDaemonClient{
		ChaosDaemonClient: chaosdaemon.NewChaosDaemonClient(cc),
		conn:              cc,
	}, nil
}
