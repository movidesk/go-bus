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

type Meta map[string]interface{}
type Header map[string]interface{}
type Body interface{}
type Message struct {
	Meta   Meta
	Header Header
	Body   Body
}
