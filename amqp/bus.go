package amqp

import (
	base "github.com/movidesk/go-bus"
)

type bus struct {
	_    struct{}
	sess *session
}

func NewBus(fns ...OptionsFn) (base.Bus, error) {
	sess, err := NewSession(fns...)
	if err != nil {
		return nil, err
	}

	return &bus{sess: sess}, nil
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
