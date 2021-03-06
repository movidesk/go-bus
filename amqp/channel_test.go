package amqp

import (
	"sync"
	"testing"
	"time"

	toxi "github.com/shopify/toxiproxy/client"
	"github.com/streadway/amqp"
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
	assert := s.Assert()

	conn, err := NewConnection(
		SetConnectionDSN("amqp://guest:guest@localhost:5672"),
	)
	assert.NoError(err)
	assert.NotNil(conn)

	chnn, err := NewChannel(conn)
	assert.NoError(err)
	assert.NotNil(chnn)
}

func (s *ChannelIntegrationSuite) TestMustChannel() {
	assert := s.Assert()

	assert.NotPanics(func() {
		conn, err := NewConnection(
			SetConnectionDSN("amqp://guest:guest@localhost:5672"),
		)
		assert.NoError(err)
		assert.NotNil(conn)

		chnn := MustChannel(conn)
		assert.NotNil(chnn)
	})
}

func (s *ChannelIntegrationSuite) TestNewChannelWithoutConfiguration() {
	assert := s.Assert()

	conn, err := NewConnection()
	assert.NoError(err)
	assert.NotNil(conn)

	chnn, err := NewChannel(conn)
	assert.NoError(err)
	assert.NotNil(chnn)
}

func (s *ChannelIntegrationSuite) TestNewChannelWithProxy() {
	assert := s.Assert()
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
	assert := s.Assert()
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
	assert := s.Assert()

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
	assert := s.Assert()

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

func (s *ChannelIntegrationSuite) TestChannelConsume() {
	assert := s.Assert()

	conn, err := NewConnection()
	assert.NoError(err)
	assert.NotNil(conn)

	chnn, err := NewChannel(conn)
	assert.NoError(err)
	assert.NotNil(chnn)

	declareTopic("amqp://guest:guest@localhost:5672", "exchange-a", "queue-a")
	defer deleteTopic("amqp://guest:guest@localhost:5672", "exchange-a", "queue-a")

	err = chnn.Publish("exchange-a", "", false, false, amqp.Publishing{Body: []byte("body")})
	assert.NoError(err)

	deliveries, err := chnn.Consume("queue-a", "", false, false, false, false, nil)
	assert.NoError(err)
	msg := <-deliveries
	msg.Ack(false)

	assert.Equal("body", string(msg.Body))

	waitToBeTrue(func() bool { return !chnn.IsClosed() }, time.Second)
	assert.False(chnn.IsClosed())
}

func (s *ChannelIntegrationSuite) TestChannelConsumeOnNetworkFailure() {
	assert := s.Assert()
	s.rabbit.Enable()
	defer s.rabbit.Disable()

	conn, err := NewConnection(
		SetConnectionDSN("amqp://guest:guest@localhost:35672"),
	)
	assert.NoError(err)
	assert.NotNil(conn)
	assert.True(!conn.IsClosed())

	chnn, err := NewChannel(
		conn,
		SetChannelPrefetchCount(2),
	)
	assert.NoError(err)
	assert.NotNil(chnn)

	declareTopic("amqp://guest:guest@localhost:5672", "exchange-b", "queue-b")
	defer deleteTopic("amqp://guest:guest@localhost:5672", "exchange-b", "queue-b")

	err = chnn.Publish("exchange-b", "", false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Body:         []byte("body 1")},
	)
	assert.NoError(err)
	err = chnn.Publish("exchange-b", "", false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Body:         []byte("body 2")},
	)
	assert.NoError(err)

	deliveries, err := chnn.Consume("queue-b", "", false, false, false, false, nil)
	assert.NoError(err)

	msg := <-deliveries
	err = msg.Nack(false, true)
	assert.NoError(err)
	assert.Equal("body 1", string(msg.Body))

	s.rabbit.Disable()
	waitToBeTrue(func() bool { return conn.IsClosed() }, time.Second)
	assert.True(conn.IsClosed())
	waitToBeTrue(func() bool { return chnn.IsClosed() }, time.Second)
	assert.True(chnn.IsClosed())

	s.rabbit.Enable()
	waitToBeTrue(func() bool { return !conn.IsClosed() }, time.Second*2)
	assert.True(!conn.IsClosed())
	waitToBeTrue(func() bool { return !chnn.IsClosed() }, time.Second*2)
	assert.True(!chnn.IsClosed())

	msg = <-deliveries
	// after failure an in flight msg channel/conneciton instance is closed
	// using msg.Ack should return an error here
	err = msg.Ack(false)
	assert.Error(err) // channel/connection is not open
	// the only way to Ack an in flight message is using the channel Acknowledgement system
	err = chnn.Ack(msg.DeliveryTag, false)
	assert.NoError(err)
	assert.Equal("body 2", string(msg.Body))

	// next message should be ok then, but avoiding msg.Ack should be a must
	msg = <-deliveries
	err = msg.Ack(false)
	assert.NoError(err)
	assert.Equal("body 1", string(msg.Body))
}

func TestChannelIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ChannelIntegrationSuite))
}
