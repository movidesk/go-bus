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

func (s *PublisherIntegrationSuite) TestPublishWithConfirm() {
	assert := assert.New(s.T())

	conn, _ := NewConnection()
	sess, _ := NewSession(conn)
	pub, _ := NewPublisher(sess)

	err, ok := pub.Publish(base.Message{})
	assert.NoError(err)
	assert.True(ok)
}

func TestPublisherIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PublisherIntegrationSuite))
}
