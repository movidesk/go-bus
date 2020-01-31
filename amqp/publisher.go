package amqp

import (
	base "github.com/movidesk/go-bus"
)

type pub struct {
}

func NewPublisher() (base.Publisher, error) {
	return &pub{}, nil
}

func (p *pub) Publish(base.Message) (error, bool) {
	return nil, false
}
