package amqp

import (
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type session struct {
	_   struct{}
	dsn string

	reconectInterval time.Duration

	locker sync.Mutex
	conn   *amqp.Connection

	terminate  chan bool
	terminated chan bool
}

func NewSession(fns ...OptionsFn) (*session, error) {
	var o Options
	o.Dsn = "amqp://guest:guest@localhost:5672"
	o.ReconectInterval = time.Millisecond * 500

	for _, fn := range fns {
		fn(&o)
	}

	sess := &session{
		dsn:              o.Dsn,
		reconectInterval: o.ReconectInterval,
	}

	err := sess.connect(sess.dial())
	if err != nil {
		return nil, err
	}
	go sess.reconnect()
	return sess, err
}

func (s *session) dial() (*amqp.Connection, error) {
	conn, err := amqp.Dial(s.dsn)
	return conn, errors.Wrap(err, "Could not amqp.Dial(dsn)")
}

func (s *session) connect(conn *amqp.Connection, err error) error {
	s.locker.Lock()
	defer s.locker.Unlock()

	s.conn = conn

	return err
}

func (s *session) reconnect() {
	for {
		reason, ok := <-s.conn.NotifyClose(make(chan *amqp.Error))
		if !ok {
			log.Println("connection closed")
			break
		}
		log.Printf("connection closed:%v", reason)

		for {
			err := s.connect(s.dial())
			if err != nil {
				time.Sleep(s.reconectInterval)
				continue
			}
			break
		}
	}
}

func (s *session) IsConnected() bool {
	s.locker.Lock()
	defer s.locker.Unlock()
	if s.conn == nil {
		return false
	}

	return !s.conn.IsClosed()
}

func (s *session) Close() {
	s.locker.Lock()
	defer s.locker.Unlock()
}
