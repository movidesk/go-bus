package amqp

import (
	"sync"
	"testing"
	"time"

	toxi "github.com/shopify/toxiproxy/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConnectionIntegrationSuite struct {
	suite.Suite
	rabbit *toxi.Proxy
}

func (s *ConnectionIntegrationSuite) SetupTest() {
	cli := toxi.NewClient("localhost:8474")
	rabbit, err := cli.Proxy("rabbit")
	if err != nil {
		rabbit, err = cli.CreateProxy("rabbit", ":35672", "mq:5672")
	}
	s.rabbit = rabbit
}

func (s *ConnectionIntegrationSuite) TestNewConnection() {
	assert := assert.New(s.T())

	conn, err := NewConnection(
		SetDSN("amqp://guest:guest@localhost:5672"),
	)

	assert.NoError(err)
	assert.NotNil(conn)
}

func (s *ConnectionIntegrationSuite) TestNewConnectionWithoutConfiguration() {
	assert := assert.New(s.T())

	conn, err := NewConnection()

	assert.NoError(err)
	assert.NotNil(conn)
}

func (s *ConnectionIntegrationSuite) TestNewConnectionWithInvalidDSN() {
	assert := assert.New(s.T())

	conn, err := NewConnection(
		SetDSN("amqp://guest:guest@invalid:5672"),
	)

	assert.Error(err, ConnectionError)
	assert.Nil(conn)
}

func (s *ConnectionIntegrationSuite) TestNewConnectionWithProxy() {
	assert := assert.New(s.T())
	s.rabbit.Enable()
	defer s.rabbit.Disable()

	conn, err := NewConnection(
		SetDSN("amqp://guest:guest@localhost:35672"),
	)
	conn.IsClosed()

	assert.NoError(err)
	assert.NotNil(conn)
	assert.False(conn.IsClosed())

	s.rabbit.Disable()
	waitToBeTrue(func() bool { return conn.IsClosed() }, time.Second)

	assert.True(conn.IsClosed())
}

func (s *ConnectionIntegrationSuite) TestConnectionOnNetworkFailure() {
	assert := assert.New(s.T())
	s.rabbit.Enable()
	defer s.rabbit.Disable()

	conn, err := NewConnection(
		SetDSN("amqp://guest:guest@localhost:35672"),
	)

	assert.NoError(err)
	assert.NotNil(conn)
	assert.True(!conn.IsClosed())

	s.rabbit.Disable()
	waitToBeTrue(func() bool { return conn.IsClosed() }, time.Second)
	assert.True(conn.IsClosed())

	s.rabbit.Enable()
	waitToBeTrue(func() bool { return !conn.IsClosed() }, time.Second)
	assert.False(conn.IsClosed())
}

func (s *ConnectionIntegrationSuite) TestConnectionOnNetworkFailureWithDelay() {
	assert := assert.New(s.T())
	s.rabbit.Enable()
	defer s.rabbit.Disable()

	conn, err := NewConnection(
		SetDSN("amqp://guest:guest@localhost:35672"),
		SetDelay(time.Second*2),
	)

	assert.NoError(err)
	assert.NotNil(conn)
	assert.True(!conn.IsClosed())

	s.rabbit.Disable()
	waitToBeTrue(func() bool { return conn.IsClosed() }, time.Second)
	assert.True(conn.IsClosed())

	s.rabbit.Enable()
	waitToBeTrue(func() bool { return !conn.IsClosed() }, time.Second)
	assert.True(conn.IsClosed())

	waitToBeTrue(func() bool { return !conn.IsClosed() }, time.Second*2)
	assert.False(conn.IsClosed())
}

func (s *ConnectionIntegrationSuite) TestConnectionWaitGroupOnConnectionClose() {
	assert := assert.New(s.T())

	wg := &sync.WaitGroup{}
	conn, err := NewConnection(
		SetDSN("amqp://guest:guest@localhost:5672"),
		SetWaitGroup(wg),
	)

	assert.NoError(err)
	assert.NotNil(conn)
	assert.True(!conn.IsClosed())

	assert.True(waitForTimeout(wg.Wait, time.Second))
	conn.Close()
	assert.False(waitForTimeout(wg.Wait, time.Second))
}

func TestConnectionIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ConnectionIntegrationSuite))
}
