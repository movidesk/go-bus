package amqp

import "github.com/pkg/errors"

type Session struct {
	*Connection
	*Channel
}

func NewSession(conn *Connection, fns ...ChannelOptionsFn) (*Session, error) {
	chnn, err := NewChannel(conn, fns...)
	if err != nil {
		return nil, err
	}
	return &Session{
		Connection: conn,
		Channel:    chnn,
	}, nil
}

func (s *Session) IsClosed() bool {
	return s.Connection.IsClosed() && s.Channel.IsClosed()
}

func (s *Session) Close() error {
	erra := s.Connection.Close()
	errb := s.Channel.Close()
	if erra == nil && errb == nil {
		return nil
	}
	return errors.Errorf("(%s|%s)", erra, errb)
}
