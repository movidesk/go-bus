package amqp

import "github.com/streadway/amqp"

type Message struct {
	*amqp.Delivery

	Headers map[string]interface{}
	Body    []byte
}

func (m *Message) Ack(multiple bool) {
	if m.Delivery == nil {
		return
	}

	m.Delivery.Ack(multiple)
}

func (m *Message) Nack(multiple bool, requeue bool) {
	if m.Delivery == nil {
		return
	}

	m.Delivery.Nack(multiple, requeue)
}

func (m *Message) Reject(requeue bool) {
	if m.Delivery == nil {
		return
	}

	m.Delivery.Reject(requeue)
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
