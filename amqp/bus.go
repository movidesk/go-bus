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
	o := &BusOptions{}
	SetBusDSN("amqp://guest:guest@localhost:5672")
	for _, fn := range fns {
		fn(o)
	}
	wg := &sync.WaitGroup{}
	close := make(chan struct{})
	return &bus{
		BusOptions: o,
		wg:         wg,
		close:      close,
	}, nil
}

func (b *bus) NewPublisher(fns ...PublisherOptionsFn) (base.Publisher, error) {
	fns = append(
		fns,
		SetPublisherClose(b.close),
		SetPublisherWaitGroup(b.wg),
	)

	if err := b.connectPub(); err != nil {
		return nil, err
	}

	sess, err := NewSession(
		b.pubconn,
		SetChannelDone(b.close),
		SetChannelWaitGroup(b.wg),
	)
	if err != nil {
		return nil, err
	}

	return NewPublisher(sess, fns...)
}

func (b *bus) NewSubscriber(fns ...SubscriberOptionsFn) (base.Subscriber, error) {
	fns = append(
		fns,
		SetSubscriberClose(b.close),
		SetSubscriberWaitGroup(b.wg),
	)

	if err := b.connectSub(); err != nil {
		return nil, err
	}

	sess, err := NewSession(
		b.subconn,
		SetChannelDone(b.close),
		SetChannelWaitGroup(b.wg),
	)
	if err != nil {
		return nil, err
	}

	return NewSubscriber(sess, fns...)
}

func (b *bus) Close() {
	close(b.close)
}

func (b *bus) Wait() {
	b.wg.Wait()
}

func (b *bus) connectPub() error {
	if b.pubconn == nil || b.pubconn.IsClosed() {
		conn, err := NewConnection(
			SetConnectionDSN(b.dsn),
			SetConnectionDone(b.close),
			SetConnectionWaitGroup(b.wg),
		)
		if err != nil {
			return err
		}
		b.pubconn = conn
	}
	return nil
}

func (b *bus) connectSub() error {
	if b.subconn == nil || b.subconn.IsClosed() {
		conn, err := NewConnection(
			SetConnectionDSN(b.dsn),
			SetConnectionDone(b.close),
			SetConnectionWaitGroup(b.wg),
		)
		if err != nil {
			return err
		}
		b.subconn = conn
	}
	return nil
}
