package fstorage_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/sptGabriel/defaults/testutils"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/network"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func TestFileStorageOperations(t *testing.T) {
	suite.Run(t, new(fileStorageTestSuite))
}

type fileStorageTestSuite struct {
	suite.Suite
	instance         *testutils.MotoInstance
	teardownInstance func()
	logger           *testutils.TestLogger
	tracer           trace.Tracer
}

func (s *fileStorageTestSuite) SetupSuite() {
	s.logger = testutils.NewTestLogger()

	ctx := context.Background()
	network, err := network.New(ctx)
	s.Require().NoError(err)

	s.instance, s.teardownInstance = testutils.SetupMoto(
		testutils.GetMotoImageName(),
		"moto-ecosystem-defaults",
		network,
		s.logger.Logger,
	)
}

func (s *fileStorageTestSuite) TearDownSuite() {
	s.teardownInstance()
}

func (s *fileStorageTestSuite) SetupTest() {
	randomId := testutils.RandomLowercaseStringRunes(10)
	s.tracer = otel.Tracer(fmt.Sprintf("test-tracer-%s", randomId))
}

func (s *fileStorageTestSuite) TearDownTest() {}
