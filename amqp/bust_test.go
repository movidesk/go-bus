package amqp

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type BusSuite struct {
	suite.Suite
}

type BusUnitSuite struct {
	BusSuite
}

type BusIntegrationSuite struct {
	BusSuite
}

func TestBusUnitSuite(t *testing.T) {
	suite.Run(t, new(BusUnitSuite))
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(BusIntegrationSuite))
}
