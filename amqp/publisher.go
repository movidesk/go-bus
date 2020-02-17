package amqp

import (
	"log"
	"sync"

	base "github.com/movidesk/go-bus"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type PublisherOptionsFn func(*PublisherOptions)

type PublisherOptions struct {
	confirm       bool
	confirmations chan amqp.Confirmation

	exchange string
	key      string

	mandatory bool
	immediate bool

	close <-chan struct{}
	wg    *sync.WaitGroup
}

func SetPublisherConfirm(confirm bool) PublisherOptionsFn {
	return func(o *PublisherOptions) {
		o.confirm = confirm
	}
}

func SetPublisherConfirmations(confirmations chan amqp.Confirmation) PublisherOptionsFn {
	return func(o *PublisherOptions) {
		o.confirmations = confirmations
	}
}

// SetPublisherExchange specifies the name of the exchange to publish to. The exchange name can be empty, meaning the default exchange. If the exchange name is specified, and that exchange does not exist, the server will raise a channel exception.
func SetPublisherExchange(exchange string) PublisherOptionsFn {
	return func(o *PublisherOptions) {
		o.exchange = exchange
	}
}

// SetPublisherKey specifies the routing key for the message. The routing key is used for routing messages depending on the exchange configuration.
func SetPublisherKey(key string) PublisherOptionsFn {
	return func(o *PublisherOptions) {
		o.key = key
	}
}

// SetPublisherMandatory this flag tells the server how to react if the message cannot be routed to a queue. If this flag is set, the server will return an unroutable message with a Return method. If this flag is zero, the server silently drops the message.
func SetPublisherMandatory(mandatory bool) PublisherOptionsFn {
	return func(o *PublisherOptions) {
		o.mandatory = mandatory
	}
}

// SetPublisherImmediate this flag tells the server how to react if the message cannot be routed to a queue consumer immediately. If this flag is set, the server will return an undeliverable message with a Return method.
func SetPublisherImmediate(immediate bool) PublisherOptionsFn {
	return func(o *PublisherOptions) {
		o.immediate = immediate
	}
}

func SetPublisherClose(close <-chan struct{}) PublisherOptionsFn {
	return func(o *PublisherOptions) {
		o.close = close
	}
}

func SetPublisherWaitGroup(wg *sync.WaitGroup) PublisherOptionsFn {
	return func(o *PublisherOptions) {
		o.wg = wg
	}
}

type Publisher interface {
	base.Publisher
}
type pub struct {
	*PublisherOptions
	*Session

	reconnected chan bool
}

func NewPublisher(sess *Session, fns ...PublisherOptionsFn) (Publisher, error) {
	o := &PublisherOptions{}
	wg := &sync.WaitGroup{}
	SetPublisherWaitGroup(wg)(o)
	SetPublisherMandatory(false)(o)
	SetPublisherImmediate(false)(o)
	SetPublisherConfirm(true)(o)
	for _, fn := range fns {
		fn(o)
	}

	reconnected := make(chan bool)
	sess.Reconnected(reconnected)
	p := &pub{
		Session:          sess,
		PublisherOptions: o,
		reconnected:      reconnected,
	}

	err := p.setup()
	if err != nil {
		return p, err
	}

	p.wg.Add(1)
	go p.loop()

	return p, err
}

func (p *pub) Publish(msg base.Message) (error, bool) {
	if p.Session.Channel.IsClosed() {
		return errors.New("Could not Publish, session channel is closed"), false
	}

	if p.Session.Connection.IsClosed() {
		return errors.New("Could not Publish, session connection is closed"), false
	}

	err := p.Session.Channel.Publish(p.exchange, p.key, p.mandatory, p.immediate,  amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Headers: msg.GetHeaders(),
		Body: msg.GetBody(),
	})
	if err != nil {
		return errors.Wrap(err, "Could not Publish, channel.Publish() failed"), false
	}

	//TODO: confirmation timeout
	c := <-p.confirmations
	return nil, c.Ack
}

func (p *pub) setup() error {
	p.confirmations = make(chan amqp.Confirmation, 1)
	if !p.confirm {
		close(p.confirmations)
		return nil
	}

	err := p.Channel.Confirm(false)
	if err != nil {
		log.Printf("publisher confirms not supported")
		close(p.confirmations)
	} else {
		log.Printf("publisher confirms enabled")
		p.Session.Channel.NotifyPublish(p.confirmations)
	}
	return err
}

func (p *pub) loop() {
	defer p.wg.Done()

	running := true
out:
	for running {
		select {
		case <-p.close:
			running = false
			break out
		case <-p.reconnected:
			if !p.confirm {
				continue
			}

			err := p.setup()
			if err != nil {
				log.Println(err.Error())
			}
		}
	}
}

