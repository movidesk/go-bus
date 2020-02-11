package amqp

import (
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type ConnOptionsFn func(*ConnOptions)

type ConnOptions struct {
	dsn   string
	delay time.Duration

	wg *sync.WaitGroup
}

func SetDSN(dsn string) ConnOptionsFn {
	return func(o *ConnOptions) {
		o.dsn = dsn
	}
}

func SetDelay(delay time.Duration) ConnOptionsFn {
	return func(o *ConnOptions) {
		o.delay = delay
	}
}

func SetWaitGroup(wg *sync.WaitGroup) ConnOptionsFn {
	return func(o *ConnOptions) {
		o.wg = wg
	}
}

type Connection struct {
	*amqp.Connection
	*ConnOptions
}

func NewConnection(fns ...ConnOptionsFn) (*Connection, error) {
	o := &ConnOptions{
		wg:  &sync.WaitGroup{},
		dsn: "amqp://guest:guest@localhost:5672",
	}
	for _, fn := range fns {
		fn(o)
	}

	conn := &Connection{
		ConnOptions: o,
	}

	err := conn.dial()
	if err != nil {
		return nil, err
	}

	go conn.loop()

	return conn, nil
}

func (c *Connection) dial() error {
	conn, err := amqp.Dial(c.dsn)
	if err != nil {
		return errors.Wrap(err, "Could not amqp.Dial(dsn)")
	}
	c.Connection = conn
	return nil
}

func (c *Connection) loop() {
	c.wg.Add(1)
	defer c.wg.Done()

	for {
		reason, ok := <-c.Connection.NotifyClose(make(chan *amqp.Error))
		if !ok {
			log.Println("connection closed")
			break
		}
		log.Printf("connection closed, reason: %v\n", reason)

		for {
			time.Sleep(c.delay)

			conn, err := amqp.Dial(c.dsn)
			if err == nil {
				c.Connection = conn
				log.Println("reconnect success")
				break
			}

			log.Printf("reconnect failed, err: %v\n", err)
		}
	}
}
