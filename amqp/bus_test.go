package amqp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BusIntegrationSuite struct {
	suite.Suite
}

func (s *BusIntegrationSuite) TestNewBus() {
	assert := assert.New(s.T())

	bus, err := NewBus()
	assert.NoError(err)
	assert.NotNil(bus)
}

func (s *BusIntegrationSuite) TestNewPublisher() {
	assert := assert.New(s.T())

	bus, err := NewBus()
	assert.NoError(err)
	assert.NotNil(bus)

	//TODO: publisher should have own settings
	pub, err := bus.NewPublisher()
	assert.NoError(err)
	assert.NotNil(pub)
}

func (s *BusIntegrationSuite) TestNewSubscriber() {
	assert := assert.New(s.T())

	bus, err := NewBus()
	assert.NoError(err)
	assert.NotNil(bus)

	//TODO: subscriber should have own settings
	sub, err := bus.NewSubscriber()
	assert.NoError(err)
	assert.NotNil(sub)
}

func TestBusIntegrationSuite(t *testing.T) {
	suite.Run(t, new(BusIntegrationSuite))
}
