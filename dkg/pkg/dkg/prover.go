package dkg

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"io/ioutil"
	"math/big"
	"path"
	"strings"
)

const (
	zokratesImage = "zokrates/zokrates"
	mountTarget   = "/home/zokrates/build"
)

type Prover struct {
	dc          *client.Client
	mountSource string
	bind        string
}

func NewProver(mountSource string) (*Prover, error) {
	dc, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	return &Prover{
		dc:          dc,
		mountSource: mountSource,
		bind:        strings.Join([]string{mountSource, mountTarget}, ":"),
	}, nil
}

func (p *Prover) ComputeWitness(ctx context.Context, args []*big.Int) error {
	var a []string
	for _, arg := range args {
		a = append(a, arg.String())
	}

	cmd := []string{
		"zokrates",
		"compute-witness",
		"-o",
		"./build/witness",
		"-i",
		"./build/out",
		"-s",
		"./build/abi.json",
		"-a",
	}
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

func (p *Prover) GenerateProof(ctx context.Context) (*Proof, error) {
	resp, err := p.dc.ContainerCreate(ctx, &container.Config{
		Image: zokratesImage,
		Cmd: []string{
			"zokrates",
			"generate-proof",
			"-i",
			"./build/out",
			"--proof-path",
			"./build/proof.json",
			"-p",
			"./build/proving.key",
			"-w",
			"./build/witness",
		},
	}, &container.HostConfig{
		Binds: []string{
			p.bind,
		},
	}, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("create container: %w", err)
	}

	if err := p.dc.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("start container: %w", err)
	}

	statusCh, errCh := p.dc.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return nil, fmt.Errorf("waiting for container: %w", err)
		}
	case <-statusCh:
	}

	if err := p.dc.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{}); err != nil {
		return nil, fmt.Errorf("remove container: %w", err)
	}

	file, err := ioutil.ReadFile(path.Join(p.mountSource, "proof.json"))
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var proof *Proof
	if err := json.Unmarshal(file, &proof); err != nil {
		return nil, fmt.Errorf("unmarshal proof: %w", err)
	}

	return proof, nil
}
