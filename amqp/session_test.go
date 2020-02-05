package amqp

import (
	"testing"

	toxi "github.com/shopify/toxiproxy/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SessionIntegrationSuite struct {
	suite.Suite
	rabbit *toxi.Proxy
}

func (s *SessionIntegrationSuite) SetupTest() {
	cli := toxi.NewClient("localhost:8474")
	rabbit, err := cli.Proxy("rabbit")
	if err != nil {
		rabbit, err = cli.CreateProxy("rabbit", ":35672", "mq:5672")
	}
	s.rabbit = rabbit
}

func (s *SessionIntegrationSuite) TestNewSession() {
	assert := assert.New(s.T())

	sess, err := NewSession(
		SetDsn("amqp://guest:guest@localhost:5672"),
	)

	assert.NoError(err)
	assert.NotNil(sess)
}

func (s *SessionIntegrationSuite) TestNewSessionWithInvalidDsn() {
	assert := assert.New(s.T())

	sess, err := NewSession(
		SetDsn("amqp://guest:guest@invalid:5672"),
	)

	assert.Error(err, ConnectionError)
	assert.Nil(sess)
}

func (s *SessionIntegrationSuite) TestNewSessionWithProxy() {
	assert := assert.New(s.T())
	s.rabbit.Enable()
	defer s.rabbit.Disable()

	sess, err := NewSession(
		SetDsn("amqp://guest:guest@localhost:35672"),
	)

	assert.NoError(err)
	assert.NotNil(sess)
	assert.True(sess.IsConnected())

	s.rabbit.Disable()
	for sess.IsConnected() {
	}

	assert.False(sess.IsConnected())
}

func (s *SessionIntegrationSuite) TestReconnectionOnNetworkFailure() {
	assert := assert.New(s.T())
	s.rabbit.Enable()
	defer s.rabbit.Disable()

	sess, err := NewSession(
		SetDsn("amqp://guest:guest@localhost:35672"),
	)

	assert.NoError(err)
	assert.NotNil(sess)
	assert.True(sess.IsConnected())

	s.rabbit.Disable()
	for sess.IsConnected() {
	}

	assert.False(sess.IsConnected())

	s.rabbit.Enable()
	for !sess.IsConnected() {
	}

	assert.True(sess.IsConnected())
}

func TestSessionIntegrationSuite(t *testing.T) {
	suite.Run(t, new(SessionIntegrationSuite))
}
