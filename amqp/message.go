package amqp

import (
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type Message struct {
	*amqp.Delivery

	Headers map[string]interface{}
	Body    []byte
}

func (m *Message) Ack(multiple bool) error {
	if m.Delivery == nil {
		return errors.New("Unable to ack message, delivery is not set")
	}

	return m.Delivery.Ack(multiple)
}

func (m *Message) Nack(multiple bool, requeue bool) error {
	if m.Delivery == nil {
		return errors.New("Unable to nack message, delivery is not set")
	}

	return m.Delivery.Nack(multiple, requeue)
}

func (m *Message) Reject(requeue bool) error {
	if m.Delivery == nil {
		return errors.New("Unable to reject message, delivery is not set")
	}

	return m.Delivery.Reject(requeue)
}

func (m *Message) GetHeaders() map[string]interface{} {
	return m.Headers
}

func (m *Message) SetHeaders(h map[string]interface{}) {
	m.Headers = h
}

func (m *Message) GetBody() []byte {
	return m.Body
}

func (m *Message) SetBody(b []byte) {
	m.Body = b
}
