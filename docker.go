package test

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/docker/docker/api/types"
	containerType "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func dockerClient(t *testing.T) *client.Client {
	cli, err := client.NewClientWithOpts()
	if err != nil {
		t.Fatalf("failed to obtain Docker client (%v)", err)
	}
	cli.NegotiateAPIVersion(context.Background())
	return cli
}

func runContainer(
	t *testing.T,
	image string,
	cmd []string,
	env []string,
	ports map[int]int,
) container {
	t.Logf("Creating and starting a container from %s...", image)
	cnt := &dockerContainer{
		client: dockerClient(t),
		image: image,
		t: t,
		cmd: cmd,
		env: env,
		ports: ports,
	}

	cnt.pull()
	cnt.create()
	t.Cleanup(cnt.remove)
	cnt.start()

	return cnt
}

type container interface {
	id() string
}

type dockerContainer struct {
	t           *testing.T
	client      *client.Client

	containerID string
	image       string
	cmd         []string
	env         []string
	ports       map[int]int
}

type nullWriter struct {

}

func (n2 nullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (d *dockerContainer) id() string {
	return d.containerID
}

func (d *dockerContainer) pull() {
	d.t.Logf("Pulling image %s...", d.image)
	reader, err := d.client.ImagePull(context.Background(), d.image, types.ImagePullOptions{})
	if err != nil {
		d.t.Fatalf("failed to pull container image %s (%v)", d.image, err)
	}
	if _, err := io.Copy(&nullWriter{}, reader); err != nil {
		d.t.Fatalf( "failed to stream logs from Minio image pull (%v)", err)
	}
}

func (d *dockerContainer) create() {
	d.t.Logf("Creating container from %s...", d.image)
	hostConfig := &containerType.HostConfig{
		AutoRemove: true,
		PortBindings: map[nat.Port][]nat.PortBinding{},
	}
	for containerPort, hostPort := range d.ports {
		portString := nat.Port(fmt.Sprintf("%d/tcp", containerPort))
		hostConfig.PortBindings[portString] = []nat.PortBinding{
			{
				HostIP: "127.0.0.1",
				HostPort: fmt.Sprintf("%d", hostPort),
			},
		}
	}
	resp, err := d.client.ContainerCreate(
		context.Background(),
		&containerType.Config{
			Image: d.image,
			Cmd:   d.cmd,
			Env:   d.env,
		},
		hostConfig,
		nil,
		nil,
		"",
	)
	if err != nil {
		d.t.Fatalf("failed to create %s container (%v)", d.image, err)
	}
	d.containerID = resp.ID
	d.t.Logf("Container has ID %s...", d.containerID)
}

func (d *dockerContainer) start() {
	d.t.Logf("Starting container %s...", d.containerID)
	if err := d.client.ContainerStart(context.Background(), d.containerID, types.ContainerStartOptions{}); err != nil {
		d.t.Fatalf("failed to start container %s (%v)", d.containerID, err)
	}
}

func (d *dockerContainer) remove() {
	d.t.Logf("Removing container ID %s...", d.containerID)
	if err := d.client.ContainerRemove(context.Background(), d.containerID, types.ContainerRemoveOptions{
		RemoveVolumes: false,
		RemoveLinks:   false,
		Force:         true,
	}); err != nil {
		d.t.Fatalf("failed to remove container %s (%v)", d.containerID, err)
	}
}
