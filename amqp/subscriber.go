package amqp

import (
	"sync"

	base "github.com/movidesk/go-bus"
)

type SubscriberOptionsFn func(*SubscriberOptions)

type SubscriberOptions struct {
	done <-chan struct{}
	wg   *sync.WaitGroup
}

type Subscriber interface {
	base.Subscriber
}

type sub struct {
}

func NewSubscriber() (Subscriber, error) {
	return &sub{}, nil
}

func (s *sub) Consume() (<-chan base.Message, <-chan struct{}, error) {
	return nil, nil, nil
}

func (s *sub) Close() {
}
