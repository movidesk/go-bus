package amqp

import (
	base "github.com/movidesk/go-bus"
	"sync"
)

type Bus interface {
	base.Bus

	NewPublisher(fns ...ChannelOptionsFn) (base.Publisher, error)
	NewSubscriber(fns ...ChannelOptionsFn) (base.Subscriber, error)
}
type bus struct {
	_       struct{}
	optconn []ConnectionOptionsFn

	pubconn *Connection
	subconn *Connection

	close chan struct{}
	wg    *sync.WaitGroup
}

func NewBus(fns ...ConnectionOptionsFn) (Bus, error) {
	wg := &sync.WaitGroup{}
	fns = append(fns, SetConnectionWaitGroup(wg))
	return &bus{
		optconn: fns,

		wg: wg,
	}, nil
}

func (b *bus) NewPublisher(fns ...ChannelOptionsFn) (base.Publisher, error) {
	if err := b.connectPub(); err != nil {
		return nil, err
	}
	fns = append(fns, SetChannelWaitGroup(b.wg))
	_, err := NewSession(b.pubconn, fns...)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (b *bus) NewSubscriber(fns ...ChannelOptionsFn) (base.Subscriber, error) {
	if err := b.connectSub(); err != nil {
		return nil, err
	}
	_, err := NewSession(b.pubconn, fns...)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (b *bus) Close() {
	b.close <- struct{}{}
}

func (b *bus) Wait() {
	b.wg.Wait()
}

func (b *bus) connectPub() error {
	if b.pubconn == nil || b.pubconn.IsClosed() {
		conn, err := NewConnection(b.optconn...)
		if err != nil {
			return err
		}
		b.pubconn = conn
	}
	return nil
}

func (b *bus) connectSub() error {
	if b.subconn == nil || b.subconn.IsClosed() {
		conn, err := NewConnection(b.optconn...)
		if err != nil {
			return err
		}
		b.subconn = conn
	}
	return nil
}
