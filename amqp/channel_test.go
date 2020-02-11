package amqp

import (
	"sync"
	"testing"
	"time"

	toxi "github.com/shopify/toxiproxy/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ChannelIntegrationSuite struct {
	suite.Suite
	rabbit *toxi.Proxy
}

func (s *ChannelIntegrationSuite) SetupTest() {
	cli := toxi.NewClient("localhost:8474")
	rabbit, err := cli.Proxy("rabbit")
	if err != nil {
		rabbit, err = cli.CreateProxy("rabbit", ":35672", "mq:5672")
	}
	s.rabbit = rabbit
}

func (s *ChannelIntegrationSuite) TestNewChannel() {
	assert := assert.New(s.T())

	conn, err := NewConnection(
		SetConnectionDSN("amqp://guest:guest@localhost:5672"),
	)
	assert.NoError(err)
	assert.NotNil(conn)

	chnn, err := NewChannel(conn)
	assert.NoError(err)
	assert.NotNil(chnn)
}

func (s *ChannelIntegrationSuite) TestNewChannelWithoutConfiguration() {
	assert := assert.New(s.T())

	conn, err := NewConnection()
	assert.NoError(err)
	assert.NotNil(conn)

	chnn, err := NewChannel(conn)
	assert.NoError(err)
	assert.NotNil(chnn)
}

func (s *ChannelIntegrationSuite) TestNewChannelWithProxy() {
	assert := assert.New(s.T())
	s.rabbit.Enable()
	defer s.rabbit.Disable()

	conn, err := NewConnection(
		SetConnectionDSN("amqp://guest:guest@localhost:35672"),
		SetConnectionDelay(time.Second*2),
	)
	assert.NoError(err)
	assert.NotNil(conn)
	assert.False(conn.IsClosed())

	chnn, err := NewChannel(conn,
		SetChannelDelay(time.Second*1),
	)
	assert.NoError(err)
	assert.NotNil(chnn)
	assert.False(chnn.IsClosed())

	s.rabbit.Disable()

	waitToBeTrue(func() bool { return conn.IsClosed() }, time.Second)
	assert.True(conn.IsClosed())

	waitToBeTrue(func() bool { return chnn.IsClosed() }, time.Second)
	assert.True(chnn.IsClosed())
}

func (s *ChannelIntegrationSuite) TestChannelOnNetworkFailure() {
	assert := assert.New(s.T())
	s.rabbit.Enable()
	defer s.rabbit.Disable()

	conn, err := NewConnection(
		SetConnectionDSN("amqp://guest:guest@localhost:35672"),
	)
	assert.NoError(err)
	assert.NotNil(conn)
	assert.True(!conn.IsClosed())

	chnn, err := NewChannel(conn)
	assert.NoError(err)
	assert.NotNil(chnn)

	s.rabbit.Disable()
	waitToBeTrue(func() bool { return conn.IsClosed() }, time.Second)
	assert.True(conn.IsClosed())
	waitToBeTrue(func() bool { return chnn.IsClosed() }, time.Second)
	assert.True(chnn.IsClosed())

	s.rabbit.Enable()
	waitToBeTrue(func() bool { return !conn.IsClosed() }, time.Second*2)
	assert.False(conn.IsClosed())
	waitToBeTrue(func() bool { return !chnn.IsClosed() }, time.Second*2)
	assert.False(chnn.IsClosed())
}

func (s *ChannelIntegrationSuite) TestChannelWaitGroupOnClose() {
	assert := assert.New(s.T())

	conn, err := NewConnection()
	assert.NoError(err)
	assert.NotNil(conn)

	wg := &sync.WaitGroup{}
	chnn, err := NewChannel(conn, SetChannelWaitGroup(wg))
	assert.NoError(err)
	assert.NotNil(wg)

	waitToBeTrue(func() bool { return !chnn.IsClosed() }, time.Second)
	assert.True(!chnn.IsClosed())

	assert.True(waitForTimeout(wg.Wait, time.Second))
	chnn.Close()
	assert.False(waitForTimeout(wg.Wait, time.Second))
}

func (s *ChannelIntegrationSuite) TestChannelWaitGroupOnDone() {
	assert := assert.New(s.T())

	conn, err := NewConnection()
	assert.NoError(err)
	assert.NotNil(conn)

	wg := &sync.WaitGroup{}
	done := make(chan struct{})
	chnn, err := NewChannel(conn,
		SetChannelWaitGroup(wg),
		SetChannelDone(done),
	)
	assert.NoError(err)
	assert.NotNil(wg)

	waitToBeTrue(func() bool { return !chnn.IsClosed() }, time.Second)
	assert.False(chnn.IsClosed())

	assert.True(waitForTimeout(wg.Wait, time.Second))
	done <- struct{}{}
	assert.False(waitForTimeout(wg.Wait, time.Second))
}

func TestChannelIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ChannelIntegrationSuite))
}
