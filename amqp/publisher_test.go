package amqp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PublisherIntegrationSuite struct {
	suite.Suite
}

func (s *PublisherIntegrationSuite) TestNewPublisher() {
	assert := s.Assert()

	conn, _ := NewConnection()
	sess, _ := NewSession(conn)
	pub, err := NewPublisher(sess)

	assert.NoError(err)
	assert.NotNil(pub)
}

func (s *PublisherIntegrationSuite) TestPublishWithConfirmOnUnexistentExchange() {
	assert := assert.New(s.T())

	conn, _ := NewConnection()
	sess, _ := NewSession(conn)
	pub, _ := NewPublisher(
		sess,
		SetPublisherExchange("unexistent"),
	)

	err, ok := pub.Publish(&Message{})
	assert.NoError(err)
	assert.False(ok)
}

func (s *PublisherIntegrationSuite) TestPublishWithConfirmOnExistentExchange() {
	assert := assert.New(s.T())

	conn, _ := NewConnection()
	sess, _ := NewSession(conn)
	pub, _ := NewPublisher(
		sess,
		SetPublisherExchange("amq.topic"),
	)

	err, ok := pub.Publish(&Message{})
	assert.NoError(err)
	assert.True(ok)
}

func (s *PublisherIntegrationSuite) TestPublishWithoutConfirmOnUnexistentExchange() {
	assert := assert.New(s.T())

	conn, _ := NewConnection()
	sess, _ := NewSession(conn)
	pub, _ := NewPublisher(
		sess,
		SetPublisherConfirm(false),
		SetPublisherExchange("unexistent"),
	)

	msg := &Message{
		Body: []byte("banana"),
	}
	err, ok := pub.Publish(msg)
	assert.NoError(err)
	assert.False(ok)
}

func (s *PublisherIntegrationSuite) TestPublishWithoutConfirmOnExistentExchange() {
	assert := s.Assert()

	conn, _ := NewConnection()
	sess, _ := NewSession(conn)
	pub, _ := NewPublisher(
		sess,
		SetPublisherConfirm(false),
		SetPublisherExchange("amq.topic"),
	)

	err, ok := pub.Publish(&Message{})
	assert.NoError(err)
	assert.False(ok)
}

func (s *PublisherIntegrationSuite) TestPublishWithHeaderAndBody() {
	assert := s.Assert()

	conn, _ := NewConnection()
	sess, _ := NewSession(conn)
	pub, _ := NewPublisher(
		sess,
		SetPublisherExchange("amq.topic"),
	)

	msg := &Message{
		Headers: map[string]interface{}{
			"key-string": "string",
			"key-int":    1,
			"key-bool":   true,
		},
		Body: []byte("banana"),
	}
	err, ok := pub.Publish(msg)
	assert.NoError(err)
	assert.True(ok)
}

func TestPublisherIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PublisherIntegrationSuite))
}
