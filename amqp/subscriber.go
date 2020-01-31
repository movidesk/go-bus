package amqp

import (
	base "github.com/movidesk/go-bus"
)

type sub struct {
}

func NewSubscriber() (base.Subscriber, error) {
	return &sub{}, nil
}

func (s *sub) Consume() (<-chan base.Message, <-chan struct{}, error) {
	return nil, nil, nil
}

func (s *sub) Close() {

}
