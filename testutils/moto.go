package testutils

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/sptGabriel/gb-libs/xlog"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const motoPort = "3000/tcp"

type MotoInstance struct {
	logger    xlog.Logger
	network   *testcontainers.DockerNetwork
	container testcontainers.Container
}

func (i *MotoInstance) ConnectionString(ctx context.Context) (string, error) {
	containerPort, err := i.container.MappedPort(ctx, motoPort)
	if err != nil {
		return "", err
	}

	host, err := i.container.Host(ctx)
	if err != nil {
		return "", err
	}

	connStr := fmt.Sprintf("http://%s", net.JoinHostPort(host, containerPort.Port()))
	return connStr, nil
}

func (i *MotoInstance) ExecuteCommand(cmd string) error {
	ctx := context.Background()

	cmdArr := []string{"/bin/sh", "-c", cmd}
	code, reader, err := i.container.Exec(ctx, cmdArr)
	if code != 0 {
		buf := new(strings.Builder)
		_, err := io.Copy(buf, reader)
		if err != nil {
			i.logger.Error("failed to copy from reader", "error", err)

			return err
		}

		bufStr := buf.String()
		i.logger.Error("command returned error exit code", "exit-code", code, "container-logs", bufStr)

		return err
	}
	if err != nil {
		i.logger.Error("failed execute command", "error", err)

		return err
	}

	return nil
}

var (
	motoInstance     *MotoInstance
	motoTearDownFunc func()
	motoOnce         sync.Once
)

func InitializeMotoInstance(
	containerName string,
	network *testcontainers.DockerNetwork,
	logger xlog.Logger,
) (*MotoInstance, func()) {
	motoOnce.Do(func() {
		minst, tearDownF := SetupMoto(GetMotoImageName(), containerName, network, logger)
		motoInstance = minst
		motoTearDownFunc = tearDownF
	})

	return motoInstance, motoTearDownFunc
}

func GetMotoImageName() string {
	envMoto := os.Getenv("MOTO_IMAGE")
	if envMoto != "" {
		return envMoto
	}

	return "motoserver/moto:5.1.5"
}

func SetupMoto(
	image, containerName string,
	network *testcontainers.DockerNetwork,
	logger xlog.Logger,
) (*MotoInstance, func()) {
	ctx := context.Background()
	request := testcontainers.ContainerRequest{
		Image: image,
		Name:  containerName,
		Env: map[string]string{
			"MOTO_PORT": "3000",
		},
		ExposedPorts: []string{motoPort},
		Networks:     []string{network.Name},
		WaitingFor: wait.ForAll(
			wait.ForHTTP("/moto-api").WithPort(motoPort),
		),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		Started:          true,
		ContainerRequest: request,
	}

	dbContainer, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		logger.Error("failed to initiate container", "error", err)
		panic(err)
	}

	tearDownMoto := func() {
		if err = dbContainer.Terminate(ctx); err != nil {
			logger.Error("failed to terminate cointainer", "error", err)
			panic(err)
		}
	}

	return &MotoInstance{
		logger:    logger,
		network:   network,
		container: dbContainer,
	}, tearDownMoto
}
