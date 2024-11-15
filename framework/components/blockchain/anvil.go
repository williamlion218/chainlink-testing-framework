package blockchain

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/smartcontractkit/chainlink-testing-framework/framework"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"time"
)

const (
	DefaultAnvilPrivateKey = `ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80`
)

// deployAnvil deploy foundry anvil node
func deployAnvil(in *Input) (*Output, error) {
	ctx := context.Background()
	entryPoint := []string{"anvil", "--host", "0.0.0.0", "--port", in.Port, "--chain-id", in.ChainID, "-b", "1"}
	entryPoint = append(entryPoint, in.DockerCmdParamsOverrides...)
	bindPort := fmt.Sprintf("%s/tcp", in.Port)
	containerName := framework.DefaultTCName("anvil")

	req := testcontainers.ContainerRequest{
		AlwaysPullImage: in.PullImage,
		Image:           fmt.Sprintf("%s", in.Image),
		Labels:          framework.DefaultTCLabels(),
		Name:            containerName,
		ExposedPorts:    []string{bindPort},
		HostConfigModifier: func(h *container.HostConfig) {
			h.PortBindings = framework.MapTheSamePort(bindPort)
		},
		Networks: []string{framework.DefaultNetworkName},
		NetworkAliases: map[string][]string{
			framework.DefaultNetworkName: {containerName},
		},
		WaitingFor: wait.ForListeningPort(nat.Port(in.Port)).WithStartupTimeout(10 * time.Second),
		Entrypoint: entryPoint,
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	host, err := framework.GetHost(c)
	if err != nil {
		return nil, err
	}
	mp, err := c.MappedPort(ctx, nat.Port(bindPort))
	if err != nil {
		return nil, err
	}
	return &Output{
		UseCache: true,
		ChainID:  in.ChainID,
		Nodes: []*Node{
			{
				HostWSUrl:             fmt.Sprintf("ws://%s:%s", host, mp.Port()),
				HostHTTPUrl:           fmt.Sprintf("http://%s:%s", host, mp.Port()),
				DockerInternalWSUrl:   fmt.Sprintf("ws://%s:%s", containerName, in.Port),
				DockerInternalHTTPUrl: fmt.Sprintf("http://%s:%s", containerName, in.Port),
			},
		},
	}, nil
}
