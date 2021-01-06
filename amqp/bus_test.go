package amqp

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"
)

type BusIntegrationSuite struct {
	suite.Suite

	exchange string
	queue    string
}

func (s *BusIntegrationSuite) SetupTest() {
	s.exchange = uuid.NewV4().String()
	s.queue = uuid.NewV4().String()
	declareTopic("amqp://guest:guest@localhost:5672", s.exchange, s.queue)
}

func (s *BusIntegrationSuite) TestNewBus() {
	assert := s.Assert()

	bus, err := NewBus()
	assert.NoError(err)
	assert.NotNil(bus)
}

func (s *BusIntegrationSuite) TestNewPublisher() {
	assert := s.Assert()

	bus, err := NewBus()
	assert.NoError(err)
	assert.NotNil(bus)

	pub, err := bus.NewPublisher()
	assert.NoError(err)
	assert.NotNil(pub)
}

func (s *BusIntegrationSuite) TestNewSubscriber() {
	assert := s.Assert()

	bus, err := NewBus()
	assert.NoError(err)
	assert.NotNil(bus)

	sub, err := bus.NewSubscriber()
	assert.NoError(err)
	assert.NotNil(sub)
}

func (s *BusIntegrationSuite) TestSubscribeWithAck() {
	assert := s.Assert()

	bus, err := NewBus()
	assert.NoError(err)
	assert.NotNil(bus)

	pub, err := bus.NewPublisher(
		SetPublisherExchange(s.exchange),
	)
	assert.NoError(err)
	assert.NotNil(pub)

	sub, err := bus.NewSubscriber(
		SetSubscriberQueue(s.queue),
	)
	assert.NoError(err)
	assert.NotNil(sub)

	err, ok := pub.Publish(&Message{
		Body: []byte("body"),
	})
	assert.NoError(err)
	assert.True(ok)

	msgs, closer, err := sub.Consume()
	done := false
out:
	for !done {
		select {
		case msg := <-msgs:
			msg.Ack(true)
			bus.Close()
		case <-closer:
			break out
		}
	}
}

func (s *BusIntegrationSuite) TestSubscribeWithNack() {
	assert := s.Assert()

	bus, err := NewBus()
	assert.NoError(err)
	assert.NotNil(bus)

	pub, err := bus.NewPublisher(
		SetPublisherExchange(s.exchange),
	)
	assert.NoError(err)
	assert.NotNil(pub)

	sub, err := bus.NewSubscriber(
		SetSubscriberQueue(s.queue),
	)
	assert.NoError(err)
	assert.NotNil(sub)

	err, ok := pub.Publish(&Message{
		Body: []byte("body"),
	})
	assert.NoError(err)
	assert.True(ok)

	counter := 0
	msgs, closer, err := sub.Consume()
	done := false
out:
	for !done {
		select {
		case msg := <-msgs:
			msg.Nack(false, true)
			counter++
			if counter >= 5 {
				bus.Close()
			}
		case <-closer:
			break out
		}
	}

	assert.Equal(5, counter)
}

func TestBusIntegrationSuite(t *testing.T) {
	suite.Run(t, new(BusIntegrationSuite))
}
