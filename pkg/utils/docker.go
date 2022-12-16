package utils

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type CreateContainerReq struct {
	Image  string            `json:"image_url"`
	Mounts map[string]string `json:"mounts"`
	Binds  map[string]string `json:"binds"`
	Name   string            `json:"name"`
	Env    []string          `json:"env"`
	Cmd    []string          `json:"cmd"`
}

type Docker struct {
	cli *client.Client
	ctx context.Context
}

var docker *Docker

func (docker *Docker) CreateContainer(req *CreateContainerReq) (string, error) {
	mounts := make([]mount.Mount, 10)
	for src, dst := range req.Mounts {
		mounts = append(mounts, mount.Mount{
			Source: src,
			Target: dst,
			Type:   mount.TypeBind,
		})
	}

	exposed_ports := nat.PortSet{}
	binds := nat.PortMap{}
	for dst, src := range req.Binds {
		p := nat.Port(src)
		exposed_ports[p] = struct{}{}
		binds[p] = []nat.PortBinding{
			{
				HostIP: "0.0.0.0",
				HostPort: dst,
			},
		}
	}

	body, err := docker.cli.ContainerCreate(docker.ctx, &container.Config{
		Hostname: req.Name,
		Env:      req.Env,
		Cmd:      req.Cmd,
		ExposedPorts: exposed_ports,
	}, &container.HostConfig{
		Mounts: mounts,
		PortBindings: binds,
	}, nil, nil, req.Name)

	if (err != nil) {
		return "", err
	}

	return body.ID, docker.StartContainer(body.ID)
}

func (docker *Docker) StartContainer(id string) error {
	return docker.cli.ContainerStart(docker.ctx, id, types.ContainerStartOptions{})
}

func (docker *Docker) StopContainer(id string) error {
	return docker.cli.ContainerStop(docker.ctx, id, nil)
}

func (docker *Docker) RestartContainer(id string) error {
	return docker.cli.ContainerRestart(docker.ctx, id , nil)
}

func (docker *Docker) RemoveContainer(id string) error {
	return docker.cli.ContainerRemove(docker.ctx, id, types.ContainerRemoveOptions{})
}

func GetDefaultDocker() (*Docker, error) {
	if docker != nil {
		return docker, nil
	}

	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	docker = &Docker{cli: client, ctx: context.Background()}
	return docker, nil
}
