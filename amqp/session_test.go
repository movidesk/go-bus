package amqp

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SessionIntegrationSuite struct {
	suite.Suite
}

func (s *SessionIntegrationSuite) SetupTest() {

}

func TestSessionIntegrationSuite(t *testing.T) {
	suite.Run(t, new(SessionIntegrationSuite))
}
