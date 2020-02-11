package amqp

import (
	base "github.com/movidesk/go-bus"
	"sync"
)

type BusOptionsFn func(*BusOptions)

type BusOptions struct {
	dsn string
}

func SetBusDSN(dsn string) BusOptionsFn {
	return func(o *BusOptions) {
		o.dsn = dsn
	}
}

type Bus interface {
	base.Bus

	NewPublisher(fns ...PublisherOptionsFn) (base.Publisher, error)
	NewSubscriber(fns ...SubscriberOptionsFn) (base.Subscriber, error)
}
type bus struct {
	_ struct{}
	*BusOptions

	pubconn *Connection
	subconn *Connection

	close chan struct{}
	wg    *sync.WaitGroup
}

func NewBus(fns ...BusOptionsFn) (Bus, error) {
	wg := &sync.WaitGroup{}
	close := make(chan struct{})
	return &bus{
		wg:    wg,
		close: close,
	}, nil
}

func (b *bus) NewPublisher(fns ...PublisherOptionsFn) (base.Publisher, error) {
	o := &PublisherOptions{}
	SetPublisherClose(b.close)(o)
	SetPublisherWaitGroup(b.wg)(o)
	for _, fn := range fns {
		fn(o)
	}

	if err := b.connectPub(); err != nil {
		return nil, err
	}
	sess, err := NewSession(b.pubconn)
	if err != nil {
		return nil, err
	}
	return NewPublisher(sess)
}

func (b *bus) NewSubscriber(fns ...SubscriberOptionsFn) (base.Subscriber, error) {
	if err := b.connectSub(); err != nil {
		return nil, err
	}
	_, err := NewSession(b.subconn)
	if err != nil {
		return nil, err
	}
	return NewSubscriber()
}

func (b *bus) Close() {
	b.close <- struct{}{}
}

func (b *bus) Wait() {
	b.wg.Wait()
}

func (b *bus) connectPub() error {
	if b.pubconn == nil || b.pubconn.IsClosed() {
		conn, err := NewConnection()
		if err != nil {
			return err
		}
		b.pubconn = conn
	}
	return nil
}

func (b *bus) connectSub() error {
	if b.subconn == nil || b.subconn.IsClosed() {
		conn, err := NewConnection()
		if err != nil {
			return err
		}
		b.subconn = conn
	}
	return nil
}
