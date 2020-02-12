package amqp

import (
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SubscriberIntegrationSuite struct {
	suite.Suite

	exchange string
	queue    string
}

func (s *SubscriberIntegrationSuite) SetupTest() {
	exchange := uuid.NewV4()
	queue := uuid.NewV4()

	s.exchange = exchange.String()
	s.queue = queue.String()
	declareTopic("amqp://guest:guest@localhost:5672", s.exchange, s.queue)
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

func TestSubscriberIntegrationSuite(t *testing.T) {
	suite.Run(t, new(SubscriberIntegrationSuite))
}
