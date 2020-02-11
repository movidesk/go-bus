package amqp

import (
	base "github.com/movidesk/go-bus"
)

type Publisher interface {
	base.Publisher
}
type pub struct {
}

func NewPublisher() (Publisher, error) {
	return &pub{}, nil
}

func (p *pub) Publish(base.Message) (error, bool) {
	return nil, false
}
