package amqp

import (
	"sync"

	base "github.com/movidesk/go-bus"
	"github.com/streadway/amqp"
)

type SubscriberOptionsFn func(*SubscriberOptions)

type SubscriberOptions struct {
	deliveries chan base.Message

	queue    string
	consumer string

	autoAck   bool
	exclusive bool
	noLocal   bool
	noWait    bool

	args amqp.Table

	close <-chan struct{}
	wg    *sync.WaitGroup
}

func SetSubscriberQueue(queue string) SubscriberOptionsFn {
	return func(o *SubscriberOptions) {
		o.queue = queue
	}
}

func SetSubscriberDeliveries(deliveries chan base.Message) SubscriberOptionsFn {
	return func(o *SubscriberOptions) {
		o.deliveries = deliveries
	}
}

func SetSubscriberClose(close <-chan struct{}) SubscriberOptionsFn {
	return func(o *SubscriberOptions) {
		o.close = close
	}
}

func SetSubscriberWaitGroup(wg *sync.WaitGroup) SubscriberOptionsFn {
	return func(o *SubscriberOptions) {
		o.wg = wg
	}
}

type Subscriber interface {
	base.Subscriber
}

type sub struct {
	*SubscriberOptions
	*Session
}

func NewSubscriber(sess *Session, fns ...SubscriberOptionsFn) (Subscriber, error) {
	o := &SubscriberOptions{}
	SetSubscriberDeliveries(make(chan base.Message, 1))(o)
	SetSubscriberWaitGroup(&sync.WaitGroup{})(o)
	for _, fn := range fns {
		fn(o)
	}
	return &sub{
		Session:           sess,
		SubscriberOptions: o,
	}, nil
}

func (s *sub) Consume() (<-chan base.Message, <-chan struct{}, error) {
	//TODO: garantee redeliveries
	deliveries, err := s.Channel.Consume(s.queue, s.consumer, s.autoAck, s.exclusive, s.noLocal, s.noWait, s.args)
	if err != nil {
		return s.deliveries, s.close, err
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		done := false
	out:
		for !done {
			select {
			case <-s.close:
				done = true
				break out
			case dlv := <-deliveries:
				msg := &Message{
					Delivery: &dlv,
					Body:     dlv.Body,
					Headers:  dlv.Headers,
				}
				s.deliveries <- msg
			}
		}
	}()

	return s.deliveries, s.close, nil
}

func (s *sub) Close() {
	close(s.deliveries)
}
