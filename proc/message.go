package proc

type Message struct {
	body    []byte
	headers map[string]interface{}
}

func (m *Message) Ack(multiple bool) error {
	return nil
}

func (m *Message) Nack(multiple bool, requeue bool) error {
	return nil
}

func (m *Message) Reject(requeue bool) error {
	return nil
}

func (m *Message) SetHeaders(h map[string]interface{}) {
	m.headers = h
}

func (m *Message) GetHeaders() map[string]interface{} {
	return m.headers
}

func (m *Message) SetBody(h []byte) {
	m.body = h
}

func (m *Message) GetBody() []byte {
	return m.body
}
