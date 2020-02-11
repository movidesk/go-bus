package amqp

import (
	base "github.com/movidesk/go-bus"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"sync"
)

type PublisherOptionsFn func(*PublisherOptions)

type PublisherOptions struct {
	confirms bool

	exchange string
	key      string

	mandatory bool
	immediate bool

	close <-chan struct{}
	wg    *sync.WaitGroup
}

func SetPublisherWaitGroup(wg *sync.WaitGroup) PublisherOptionsFn {
	return func(o *PublisherOptions) {
		o.wg = wg
	}
}

func SetPublisherClose(close <-chan struct{}) PublisherOptionsFn {
	return func(o *PublisherOptions) {
		o.close = close
	}
}

type Publisher interface {
	base.Publisher
}
type pub struct {
	*PublisherOptions
	*Session
}

func NewPublisher(sess *Session, fns ...PublisherOptionsFn) (Publisher, error) {
	o := &PublisherOptions{}
	for _, fn := range fns {
		fn(o)
	}

	return &pub{
		Session:          sess,
		PublisherOptions: o,
	}, nil
}

func (p *pub) Publish(base.Message) (error, bool) {
	err := p.Session.Publish(p.exchange, p.key, p.mandatory, p.immediate, amqp.Publishing{})
	if err != nil {
		return errors.Wrap(err, "Could not Session.Publish()"), false
	}
	return nil, false
}
