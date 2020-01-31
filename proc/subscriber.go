package proc

import (
	"sync"

	base "github.com/movidesk/go-bus"
)

type sub struct {
	_   struct{}
	out <-chan base.Message

	closer chan struct{}
	wg     *sync.WaitGroup
}

func NewSubscriber(fns ...OptionsFn) (base.Subscriber, error) {
	var o Options
	for _, fn := range fns {
		fn(&o)
	}

	o.wg.Add(1)

	return &sub{
		out: o.out,

		closer: o.closer,
		wg:     o.wg,
	}, nil
}

func (s *sub) Consume() (<-chan base.Message, <-chan struct{}, error) {
	return s.out, s.closer, nil
}

func (s *sub) Close() {
	s.wg.Done()
}
