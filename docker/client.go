package docker

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/Squwid/bg-compiler/flags"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var Client DockerClient

type DockerClient interface {
	CreateContainer(context.Context, *CreateContainerInput) (string, error)

	StartContainer(context.Context, string) (io.ReadCloser, error)

	KillContainer(context.Context, string) (bool, error)

	FeedStdIn(ctx context.Context, id, stdIn string) error
}

type CreateContainerInput struct {
	ID          string // ID of the JobDefinition.
	FullCommand string // Full command. Ex. "python3 main.py", "go run main.go"
	Mounts      []mount.Mount
	Image       string // Docker image (must be available on host)
}

type dclient struct {
	docker *client.Client
}

var (
	targetArc     = os.Getenv("TARGET_ARCH")
	targetOS      = os.Getenv("TARGET_OS")
	targetVariant = os.Getenv("TARGET_VARIANT") // v8 for rpi.
)

func Init() {
	if targetArc == "" {
		logrus.Warnf("'TARGET_ARCH' empty, defaulting to 'amd64'")
		targetArc = "amd64"
	}
	if targetOS == "" {
		logrus.Warnf("'TARGET_OS' empty, defaulting to 'linux'")
		targetOS = "linux"
	}

	c, err := client.NewClientWithOpts(client.FromEnv,
		client.WithVersion("1.44"))
	if err != nil {
		logrus.WithError(err).Fatalf("Error creating docker client")
	}

	Client = &dclient{docker: c}
}

func (c *dclient) CreateContainer(ctx context.Context,
	input *CreateContainerInput) (string, error) {
	hostConfig := &container.HostConfig{
		Mounts: input.Mounts,
		LogConfig: container.LogConfig{
			Type: "json-file",
			Config: map[string]string{
				"mode": "non-blocking",
			},
		},
		AutoRemove: true,
		Resources: container.Resources{
			CPUShares: flags.ContainerCPUShares(),
			Memory:    flags.ContainerMaxMemory(),
		},
		Privileged: false,
	}
	if flags.UseGVisor() {
		hostConfig.Runtime = "runsc"
	}

	resp, err := c.docker.ContainerCreate(ctx,
		&container.Config{
			OpenStdin:       true,
			Tty:             false,
			AttachStdin:     true,
			AttachStdout:    true,
			Image:           input.Image,
			NetworkDisabled: true,
			Cmd: strslice.StrSlice{
				"/bin/sh", "-c", input.FullCommand,
			},
			WorkingDir: "/bg",
		},
		hostConfig,
		&network.NetworkingConfig{},
		&v1.Platform{
			Architecture: targetArc,
			OS:           targetOS,
			Variant:      targetVariant,
		},
		input.ID)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

// StartContainer starts a container and returns a ReadCloser to the logs.
// They still need to be multiplexed between stdout and stderr.
func (c *dclient) StartContainer(ctx context.Context, id string) (io.ReadCloser, error) {
	if err := c.docker.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		return nil, err
	}

	logs, err := c.docker.ContainerLogs(context.Background(), id, container.LogsOptions{
		Timestamps: false,
		Follow:     true,
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return nil, err
	}

	return logs, nil
}

// KillContainer returns a bool whether the container was running or not (indicating a timeout)
// as well as the error.
func (c *dclient) KillContainer(ctx context.Context, id string) (bool, error) {
	status, err := c.docker.ContainerInspect(ctx, id)
	if err != nil {
		return false, errors.Wrapf(err, "Error inspecting container")
	}
	if !status.State.Running {
		return false, nil
	}
	return true, c.docker.ContainerKill(ctx, id, "SIGKILL")
}

func (c *dclient) FeedStdIn(ctx context.Context, id, stdIn string) error {
	if !strings.HasSuffix(stdIn, "\n") {
		stdIn += "\n"
	}

	resp, err := c.docker.ContainerAttach(ctx, id, container.AttachOptions{
		Stream: true,
		Stdin:  true,
	})
	if err != nil {
		return errors.Wrapf(err, "Error attatching directly to container for StdIn")
	}

	_, err = resp.Conn.Write([]byte(stdIn))
	if err != nil {
		return errors.Wrapf(err, "Error sending stdin to container")
	}

	resp.Conn.Close()
	return nil
}
