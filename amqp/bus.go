package amqp

import (
	base "github.com/movidesk/go-bus"
)

type bus struct {
	_       struct{}
	pubconn *Connection
	subconn *Connection
}

func NewBus(fns ...OptionsFn) (base.Bus, error) {
	return &bus{}, nil
}

func (b *bus) NewPublisher() (base.Publisher, error) {
	return nil, nil
}

func (b *bus) NewSubscriber() (base.Subscriber, error) {
	return nil, nil
}

func (b *bus) Close() {
}

func (b *bus) Wait() {
}
