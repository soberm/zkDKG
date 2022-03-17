package dkg

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"math/big"
	"strings"
)

const (
	zokratesImage = "zokrates/zokrates"
	mountTarget   = "/home/zokrates/build"
)

type Prover struct {
	dc   *client.Client
	bind string
}

func NewProver(mountSource string) (*Prover, error) {
	dc, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	return &Prover{
		dc:   dc,
		bind: strings.Join([]string{mountSource, mountTarget}, ":"),
	}, nil
}

func (p *Prover) ComputeWitness(ctx context.Context, args []*big.Int) error {
	var a []string
	for _, arg := range args {
		a = append(a, arg.String())
	}

	cmd := []string{"zokrates", "compute-witness", "-o", "./build/witness", "-i", "./build/out", "-s", "./build/abi.json", "-a"}
	resp, err := p.dc.ContainerCreate(ctx, &container.Config{
		Image: zokratesImage,
		Cmd:   append(cmd, a...),
	}, &container.HostConfig{
		Binds: []string{
			p.bind,
		},
	}, nil, nil, "")
	if err != nil {
		return fmt.Errorf("create container: %w", err)
	}
	if err := p.dc.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	statusCh, errCh := p.dc.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("waiting for container: %w", err)
		}
	case <-statusCh:
	}

	if err := p.dc.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{}); err != nil {
		return fmt.Errorf("remove container: %w", err)
	}
	return nil
}

func (p *Prover) GenerateProof(ctx context.Context) error {
	resp, err := p.dc.ContainerCreate(ctx, &container.Config{
		Image: zokratesImage,
		Cmd:   []string{"zokrates", "generate-proof", "-i", "./build/out", "--proof-path", "./build/proof.json", "-p", "./build/proving.key", "-w", "./build/witness"},
	}, &container.HostConfig{
		Binds: []string{
			p.bind,
		},
	}, nil, nil, "")
	if err != nil {
		return fmt.Errorf("create container: %w", err)
	}

	if err := p.dc.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	statusCh, errCh := p.dc.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("waiting for container: %w", err)
		}
	case <-statusCh:
	}

	if err := p.dc.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{}); err != nil {
		return fmt.Errorf("remove container: %w", err)
	}
	return nil
}
