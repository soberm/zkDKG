package dkg

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type ProofType string
const (
	EvalPolyProof ProofType = "poly_eval"
	KeyDerivProof ProofType = "key_deriv"

	zokratesImage = "zokrates/zokrates:0.8.2"
	mountTarget   = "/home/zokrates/build"
)

type Prover struct {
	dc          *client.Client
	mountSource string
	bind        string
	pipe		*os.File
}

func NewProver(mountSource string, pipe *os.File) (*Prover, error) {
	dc, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	return &Prover{
		dc:          dc,
		mountSource: mountSource,
		bind:        strings.Join([]string{mountSource, mountTarget}, ":"),
		pipe:		 pipe,
	}, nil
}

func (p *Prover) ComputeWitness(ctx context.Context, proofType ProofType, args []*big.Int) error {
	var a []string
	for _, arg := range args {
		a = append(a, arg.String())
	}

	basePath := path.Join("./build", string(proofType))

	cmd := []string{
		"zokrates",
		"compute-witness",
		"-o",
		path.Join(basePath, "witness"),
		"-i",
		path.Join(basePath, "out"),
		"-s",
		path.Join(basePath, "abi.json"),
		"-a",
	}

	user, err := user.Current()
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	resp, err := p.dc.ContainerCreate(ctx, &container.Config{
		Image: zokratesImage,
		User: fmt.Sprintf("%s:%s", user.Uid, user.Gid),
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
	case status := <-statusCh:
		if status.StatusCode != 0 {
			var msg string
			if status.Error == nil {
				msg = fmt.Sprintf("exit code %d", status.StatusCode)
			} else {
				msg = status.Error.Message
			}
			return fmt.Errorf("running container: %s", msg)
		}
	}

	return nil
}

func (p *Prover) GenerateProof(ctx context.Context, proofType ProofType) (*Proof, error) {
	basePath := path.Join("./build", string(proofType))
	user, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	resp, err := p.dc.ContainerCreate(ctx, &container.Config{
		Image: zokratesImage,
		User: fmt.Sprintf("%s:%s", user.Uid, user.Gid),
		Cmd: []string{
			"zokrates",
			"generate-proof",
			"-i",
			path.Join(basePath, "out"),
			"--proof-path",
			path.Join(basePath, "proof.json"),
			"-p",
			path.Join(basePath, "proving.key"),
			"-w",
			path.Join(basePath, "witness"),
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
	case status := <-statusCh:
		if status.StatusCode != 0 {
			var msg string
			if status.Error == nil {
				msg = fmt.Sprintf("exit code %d", status.StatusCode)
			} else {
				msg = status.Error.Message
			}
			return nil, fmt.Errorf("running container: %s", msg)
		}
	}

	if p.pipe != nil {
		json, err := p.dc.ContainerInspect(ctx, resp.ID)
		if err != nil {
			return nil, fmt.Errorf("inspect container: %w", err)
		}

		if _, err = p.pipe.WriteString(json.Name[1:] + "\n"); err != nil {
			return nil, fmt.Errorf("write to pipe: %w", err)
		}
	}

	file, err := ioutil.ReadFile(path.Join(p.mountSource, string(proofType), "proof.json"))
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var proof *Proof
	if err := json.Unmarshal(file, &proof); err != nil {
		return nil, fmt.Errorf("unmarshal proof: %w", err)
	}

	return proof, nil
}

func (p *Prover) Close() {
	if p.pipe != nil {
		p.pipe.Close()
	}
}
