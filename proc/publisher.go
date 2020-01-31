package proc

import (
	"errors"

	base "github.com/movidesk/go-bus"
)

type pub struct {
	_  struct{}
	in chan<- base.Message

	closer <-chan struct{}
	closed bool
}

func NewPublisher(fns ...OptionsFn) (base.Publisher, error) {
	var o Options
	for _, fn := range fns {
		fn(&o)
	}
	return &pub{
		in:     o.in,
		closer: o.closer,
	}, nil
}

func (p *pub) Publish(msg base.Message) (error, bool) {
	if p.closed {
		return errors.New("closed"), false
	}
	p.in <- msg
	return nil, true
}

func (p *pub) Close() {
	go func() {
		<-p.closer
		p.closed = true
	}()
}
