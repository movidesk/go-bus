package proc

import (
	base "github.com/movidesk/go-bus"
	"sync"
)

type OptionsFn func(*Options)

type Options struct {
	_   struct{}
	in  chan<- base.Message
	out <-chan base.Message

	closer chan struct{}
	wg     *sync.WaitGroup
}

func SetIn(in chan<- base.Message) OptionsFn {
	return func(o *Options) {
		o.in = in
	}
}

func SetOut(out <-chan base.Message) OptionsFn {
	return func(o *Options) {
		o.out = out
	}
}

func SetCloser(closer chan struct{}) OptionsFn {
	return func(o *Options) {
		o.closer = closer
	}
}

func SetWaitGroup(wg *sync.WaitGroup) OptionsFn {
	return func(o *Options) {
		o.wg = wg
	}
}
