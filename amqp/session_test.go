package amqp

import (
	"io"
	"log"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SessionIntegrationSuite struct {
	suite.Suite
}

func (s *SessionIntegrationSuite) TestNewSession() {
	assert := assert.New(s.T())

	sess, err := NewSession(
		SetDsn("amqp://guest:guest@localhost:5672"),
	)

	assert.NoError(err)
	assert.NotNil(sess)
}

func (s *SessionIntegrationSuite) TestNewSessionWithInvalidDsn() {
	assert := assert.New(s.T())

	sess, err := NewSession(
		SetDsn("amqp://guest:guest@invalid:5672"),
	)

	assert.Error(err, ConnectionError)
	assert.Nil(sess)
}

func (s *SessionIntegrationSuite) TestNewSessionWithProxy() {
	assert := assert.New(s.T())
	done := make(chan struct{})
	var wg sync.WaitGroup
	go proxy(":35672", "localhost:5672", &wg)(done)

	sess, err := NewSession(
		SetDsn("amqp://guest:guest@localhost:35672"),
	)

	assert.NoError(err)
	assert.NotNil(sess)
	assert.True(sess.IsConnected())

	done <- struct{}{}
	wg.Wait()
	for sess.IsConnected() {
	}

	assert.False(sess.IsConnected())
}

func (s *SessionIntegrationSuite) TestReconnectionOnNetworkFailure() {
	assert := assert.New(s.T())
	done := make(chan struct{})
	var wg sync.WaitGroup
	go proxy(":45672", "localhost:5672", &wg)(done)

	sess, err := NewSession(
		SetDsn("amqp://guest:guest@localhost:45672"),
	)

	assert.NoError(err)
	assert.NotNil(sess)
	assert.True(sess.IsConnected())

	done <- struct{}{}
	wg.Wait()
	for sess.IsConnected() {
	}

	assert.False(sess.IsConnected())

	go proxy(":45672", "localhost:5672", &wg)(done)
	for !sess.IsConnected() {

	}

	assert.True(sess.IsConnected())

	done <- struct{}{}
	wg.Wait()
}

func TestSessionIntegrationSuite(t *testing.T) {
	suite.Run(t, new(SessionIntegrationSuite))
}

func proxy(from, to string, wg *sync.WaitGroup) func(<-chan struct{}) {
	return func(done <-chan struct{}) {
		incoming, err := net.Listen("tcp", from)
		if err != nil {
			log.Fatalf("could not start server on %s: %v", from, err)
		}
		wg.Add(1)
		defer close(incoming.Close, wg)
		log.Printf("server running on %s\n", from)

		client, err := incoming.Accept()
		if err != nil {
			log.Fatal("could not accept client connection", err)
		}
		wg.Add(1)
		defer close(client.Close, wg)
		log.Printf("client '%v' connected!\n", client.RemoteAddr())

		target, err := net.Dial("tcp", to)
		if err != nil {
			log.Fatal("could not connect to target", err)
		}
		wg.Add(1)
		defer close(target.Close, wg)
		log.Printf("connection to server %v established!\n", target.RemoteAddr())

		go func() { io.Copy(target, client) }()
		go func() { io.Copy(client, target) }()

		<-done
		log.Println("done!")
	}
}

func close(fn func() error, wg *sync.WaitGroup) error {
	defer wg.Done()
	return fn()
}
