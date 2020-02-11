package amqp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	base "github.com/movidesk/go-bus"
)

type PublisherIntegrationSuite struct {
	suite.Suite
}

func (s *PublisherIntegrationSuite) TestNewPublisher() {
	assert := assert.New(s.T())

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

	err, ok := pub.Publish(base.Message{})
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

	err, ok := pub.Publish(base.Message{})
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

	err, ok := pub.Publish(base.Message{})
	assert.NoError(err)
	assert.False(ok)
}

func (s *PublisherIntegrationSuite) TestPublishWithoutConfirmOnExistentExchange() {
	assert := assert.New(s.T())

	conn, _ := NewConnection()
	sess, _ := NewSession(conn)
	pub, _ := NewPublisher(
		sess,
		SetPublisherConfirm(false),
		SetPublisherExchange("amq.topic"),
	)

	err, ok := pub.Publish(base.Message{})
	assert.NoError(err)
	assert.False(ok)
}

func (s *PublisherIntegrationSuite) TestPublishWithHeaderAndBody() {
	assert := assert.New(s.T())

	conn, _ := NewConnection()
	sess, _ := NewSession(conn)
	pub, _ := NewPublisher(
		sess,
		SetPublisherExchange("amq.topic"),
	)

	type Fruit struct {
		Name string `json:"name"`
	}

	err, ok := pub.Publish(base.Message{
		Header: base.Header{
			"key-string": "value",
			"key-int":    1,
			"key-bool":   true,
		},
		Body: &Fruit{"Banana"},
	})
	assert.NoError(err)
	assert.True(ok)
}

func TestPublisherIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PublisherIntegrationSuite))
}
