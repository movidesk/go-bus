package bus

type Bus interface {
	Wait()
	Close()
}

type Publisher interface {
	Publish(Message) (error, bool)
}

type Subscriber interface {
	Consume() (<-chan Message, <-chan struct{}, error)

	Close()
}

type Message interface {
	Ack(multiple bool) error
	Nack(multiple bool, requeue bool) error
	Reject(requeue bool) error

	GetHeaders() map[string]interface{}
	SetHeaders(map[string]interface{})
	GetBody() []byte
	SetBody([]byte)
}
