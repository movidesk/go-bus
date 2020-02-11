package proc

import (
	base "github.com/movidesk/go-bus"
	"sync"
)

type Bus interface {
	base.Bus
	NewPublisher() (base.Publisher, error)
	NewSubscriber() (base.Subscriber, error)
}

type bus struct {
	_   struct{}
	in  chan<- base.Message
	out <-chan base.Message

	closed bool
	closer chan struct{}
	wg     *sync.WaitGroup
}

func NewBus(fns ...OptionsFn) (Bus, error) {
	var o Options
	var wg sync.WaitGroup
	o.closer = make(chan struct{})
	o.wg = &wg

	for _, fn := range fns {
		fn(&o)
	}

	o.wg.Add(1)

	return &bus{
		in:     o.in,
		out:    o.out,
		closer: o.closer,
		wg:     o.wg,
	}, nil
}

func (b *bus) NewPublisher() (base.Publisher, error) {
	return NewPublisher(
		SetIn(b.in),

		SetCloser(b.closer),
	)
}

func (b *bus) NewSubscriber() (base.Subscriber, error) {
	return NewSubscriber(
		SetOut(b.out),

		SetCloser(b.closer),
		SetWaitGroup(b.wg),
	)
}

func (b *bus) Close() {
	b.wg.Done()
	b.closer <- struct{}{}
}

func (b *bus) Wait() {
	b.wg.Wait()
}
