package amqp

type session struct {
}

func NewSession() (*session, error) {
	return &session{}, nil
}
