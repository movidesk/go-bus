package amqp

import (
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	toxi "github.com/shopify/toxiproxy/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SubscriberIntegrationSuite struct {
	suite.Suite

	exchange string
	queue    string

	rabbit *toxi.Proxy
}

func (s *SubscriberIntegrationSuite) SetupTest() {
	exchange := uuid.NewV4()
	queue := uuid.NewV4()

	s.exchange = exchange.String()
	s.queue = queue.String()
	declareTopic("amqp://guest:guest@localhost:5672", s.exchange, s.queue)

	cli := toxi.NewClient("localhost:8474")
	rabbit, err := cli.Proxy("rabbit")
	if err != nil {
		rabbit, err = cli.CreateProxy("rabbit", ":35672", "mq:5672")
	}
	s.rabbit = rabbit

}

func (s *SubscriberIntegrationSuite) TestNewSubscriber() {
	assert := assert.New(s.T())

	conn, _ := NewConnection()
	sess, _ := NewSession(conn)

	sub, err := NewSubscriber(sess)

	assert.NoError(err)
	assert.NotNil(sub)
}

func (s *SubscriberIntegrationSuite) TestConsumeOnClose() {
	assert := assert.New(s.T())

	conn, _ := NewConnection()
	sess, _ := NewSession(conn)

	done := make(chan struct{})
	sub, err := NewSubscriber(
		sess,
		SetSubscriberClose(done),
		SetSubscriberQueue(s.queue),
	)
	assert.NoError(err)
	assert.NotNil(sub)

	_, closer, err := sub.Consume()
	assert.NoError(err)

	open := true
	t := time.NewTimer(time.Millisecond * 500)
out:
	for open {
		select {
		case <-t.C:
			t.Stop()
			close(done)
		case <-closer:
			open = false
			break out
		}
	}
	assert.False(open)
}

func (s *SubscriberIntegrationSuite) TestConsumeOnNetworkFailure() {
	assert := assert.New(s.T())
	s.rabbit.Enable()
	defer s.rabbit.Disable()

	conn, _ := NewConnection(SetConnectionDSN("amqp://guest:guest@localhost:35672"))
	sess, _ := NewSession(conn)

	done := make(chan struct{})
	pub, err := NewPublisher(
		sess,
		SetPublisherClose(done),
		SetPublisherExchange(s.exchange),
	)
	assert.NoError(err)
	assert.NotNil(pub)

	sub, err := NewSubscriber(
		sess,
		SetSubscriberClose(done),
		SetSubscriberQueue(s.queue),
	)
	assert.NoError(err)
	assert.NotNil(sub)

	err, ok := pub.Publish(&Message{
		Body: []byte("body-a"),
	})
	assert.NoError(err)
	assert.True(ok)

	err, ok = pub.Publish(&Message{
		Body: []byte("body-b"),
	})
	assert.NoError(err)
	assert.True(ok)

	deliveries, _, err := sub.Consume()
	assert.NoError(err)

	msg := <-deliveries
	err = msg.Nack(false, true)
	assert.NoError(err)
	assert.Equal("body-a", string(msg.GetBody()))

	s.rabbit.Disable()
	waitToBeTrue(func() bool { return conn.IsClosed() }, time.Second*2)
	assert.True(conn.IsClosed())

	s.rabbit.Enable()
	waitToBeTrue(func() bool { return !conn.IsClosed() }, time.Second*2)
	assert.True(!conn.IsClosed())

	//will skip a delivery without a channel

	msg = <-deliveries
	err = msg.Ack(false)
	assert.NoError(err)
	assert.Equal("body-a", string(msg.GetBody()))

	msg = <-deliveries
	err = msg.Ack(false)
	assert.NoError(err)
	assert.Equal("body-b", string(msg.GetBody()))
}

func (s *SubscriberIntegrationSuite) TestConsumeOnNetworkFailureWhileMessageIsInFlight() {
	assert := assert.New(s.T())
	s.rabbit.Enable()
	defer s.rabbit.Disable()

	conn, _ := NewConnection(SetConnectionDSN("amqp://guest:guest@localhost:35672"))
	sess, _ := NewSession(conn)

	done := make(chan struct{})
	pub, err := NewPublisher(
		sess,
		SetPublisherClose(done),
		SetPublisherExchange(s.exchange),
	)
	assert.NoError(err)
	assert.NotNil(pub)

	sub, err := NewSubscriber(
		sess,
		SetSubscriberClose(done),
		SetSubscriberQueue(s.queue),
	)
	assert.NoError(err)
	assert.NotNil(sub)

	err, ok := pub.Publish(&Message{
		Body: []byte("body"),
	})
	assert.NoError(err)
	assert.True(ok)

	deliveries, _, err := sub.Consume()
	assert.NoError(err)

	msg := <-deliveries

	s.rabbit.Disable()
	waitToBeTrue(func() bool { return conn.IsClosed() }, time.Second*2)
	assert.True(conn.IsClosed())

	err = msg.Nack(false, true)
	assert.Error(err)
	assert.Equal("body", string(msg.GetBody()))

	s.rabbit.Enable()
	waitToBeTrue(func() bool { return !conn.IsClosed() }, time.Second*2)
	assert.True(!conn.IsClosed())

	msg = <-deliveries
	err = msg.Ack(false)
	assert.NoError(err)
	assert.Equal("body", string(msg.GetBody()))
}

func TestSubscriberIntegrationSuite(t *testing.T) {
	suite.Run(t, new(SubscriberIntegrationSuite))
}
