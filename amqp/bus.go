package amqp

import (
	"context"
	"sync"

	base "github.com/movidesk/go-bus"
	"github.com/pkg/errors"
)

type BusOptionsFn func(*BusOptions)

type BusOptions struct {
	dsn string
	wg  *sync.WaitGroup
}

func SetBusDSN(dsn string) BusOptionsFn {
	return func(o *BusOptions) {
		o.dsn = dsn
	}
}

func SetBusWaitGroup(wg *sync.WaitGroup) BusOptionsFn {
	return func(o *BusOptions) {
		o.wg = wg
	}
}

type Bus interface {
	base.Bus

	NewPublisher(fns ...PublisherOptionsFn) (base.Publisher, error)
	MustPublisher(fns ...PublisherOptionsFn) base.Publisher
	NewSubscriber(fns ...SubscriberOptionsFn) (base.Subscriber, error)
	MustSubscriber(fns ...SubscriberOptionsFn) base.Subscriber
}

type bus struct {
	_ struct{}
	*BusOptions

	pubconn *Connection
	subconn *Connection

	close chan struct{}
	wg    *sync.WaitGroup
}

func MustBus(fns ...BusOptionsFn) Bus {
	bus, err := NewBus(fns...)
	if err != nil {
		panic(err)
	}
	return bus
}

func NewBus(fns ...BusOptionsFn) (Bus, error) {
	o := &BusOptions{}
	SetBusDSN("amqp://guest:guest@localhost:5672")(o)
	SetBusWaitGroup(&sync.WaitGroup{})(o)
	for _, fn := range fns {
		fn(o)
	}
	close := make(chan struct{})
	return &bus{
		BusOptions: o,
		wg:         o.wg,
		close:      close,
	}, nil
}

func (b *bus) MustPublisher(fns ...PublisherOptionsFn) base.Publisher {
	fns = append(
		fns,
		SetPublisherClose(b.close),
		SetPublisherWaitGroup(b.wg),
	)

	if err := b.connectPub(); err != nil {
		panic(err)
	}

	sess := MustSession(
		b.pubconn,
		SetChannelDone(b.close),
		SetChannelWaitGroup(b.wg),
	)

	return MustPublisher(sess, fns...)
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

func (b *bus) MustSubscriber(fns ...SubscriberOptionsFn) base.Subscriber {
	fns = append(
		fns,
		SetSubscriberClose(b.close),
		SetSubscriberWaitGroup(b.wg),
	)

	if err := b.connectSub(); err != nil {
		panic(err)
	}

	sess := MustSession(
		b.subconn,
		SetChannelDone(b.close),
		SetChannelWaitGroup(b.wg),
	)

	return MustSubscriber(sess, fns...)
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

func (b *bus) Shutdown(timeout context.Context) error {
	b.Close()
	closed := make(chan struct{})
	go func() {
		defer close(closed)
		b.Wait()
	}()

	select {
	case <-timeout.Done():
		return errors.New("closed by timeout")
	case <-closed:
		return nil
	}
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
