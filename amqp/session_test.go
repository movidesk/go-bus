package amqp

import (
	"sync"
	"testing"
	"time"

	toxi "github.com/shopify/toxiproxy/client"
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
	assert := s.Assert()

	conn, err := NewConnection(
		SetConnectionDSN("amqp://guest:guest@localhost:5672"),
	)
	assert.NoError(err)
	assert.NotNil(conn)

	sess, err := NewSession(conn)
	assert.NoError(err)
	assert.NotNil(sess)
}

func (s *SessionIntegrationSuite) TestSessionWaitGroupOnClose() {
	assert := s.Assert()

	wg := &sync.WaitGroup{}
	conn, err := NewConnection(
		SetConnectionWaitGroup(wg),
	)
	assert.NoError(err)
	assert.NotNil(conn)

	sess, err := NewSession(conn,
		SetChannelWaitGroup(wg),
	)
	assert.NoError(err)
	assert.NotNil(sess)

	waitToBeTrue(func() bool { return !sess.IsClosed() }, time.Second)
	assert.True(!sess.IsClosed())

	assert.True(waitForTimeout(wg.Wait, time.Second))
	sess.Close()
	assert.False(waitForTimeout(wg.Wait, time.Second))
}

func (s *SessionIntegrationSuite) TestSessionWaitGroupOnDone() {
	assert := s.Assert()

	wg := &sync.WaitGroup{}
	done := make(chan struct{})
	conn, err := NewConnection(
		SetConnectionWaitGroup(wg),
		SetConnectionDone(done),
	)
	assert.NoError(err)
	assert.NotNil(conn)

	sess, err := NewSession(conn,
		SetChannelWaitGroup(wg),
		SetChannelDone(done),
	)
	assert.NoError(err)
	assert.NotNil(sess)

	waitToBeTrue(func() bool { return !sess.IsClosed() }, time.Second)
	assert.False(sess.IsClosed())

	assert.True(waitForTimeout(wg.Wait, time.Second))
	done <- struct{}{}
	assert.False(waitForTimeout(wg.Wait, time.Second))
}

func TestSessionIntegrationSuite(t *testing.T) {
	suite.Run(t, new(SessionIntegrationSuite))
}
