package amqp

import (
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type ConnectionOptionsFn func(*ConnectionOptions)

type ConnectionOptions struct {
	dsn   string
	delay time.Duration

	wg   *sync.WaitGroup
	done <-chan struct{}
}

func SetConnectionDSN(dsn string) ConnectionOptionsFn {
	return func(o *ConnectionOptions) {
		o.dsn = dsn
	}
}

func SetConnectionDelay(delay time.Duration) ConnectionOptionsFn {
	return func(o *ConnectionOptions) {
		o.delay = delay
	}
}

func SetConnectionWaitGroup(wg *sync.WaitGroup) ConnectionOptionsFn {
	return func(o *ConnectionOptions) {
		o.wg = wg
	}
}

func SetConnectionDone(done <-chan struct{}) ConnectionOptionsFn {
	return func(o *ConnectionOptions) {
		o.done = done
	}
}

type Connection struct {
	*amqp.Connection
	*ConnectionOptions
}

func NewConnection(fns ...ConnectionOptionsFn) (*Connection, error) {
	o := &ConnectionOptions{
		wg:    &sync.WaitGroup{},
		dsn:   "amqp://guest:guest@localhost:5672",
		delay: time.Second,
	}
	for _, fn := range fns {
		fn(o)
	}

	conn := &Connection{
		ConnectionOptions: o,
	}

	err := conn.dial()
	if err != nil {
		return nil, err
	}

	conn.wg.Add(1)
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
	defer c.wg.Done()
	running := true

out:
	for running {
		select {
		case <-c.done:
			c.Close()

		case reason, ok := <-c.Connection.NotifyClose(make(chan *amqp.Error)):
			if !ok {
				log.Println("connection closed")
				running = false
				break out
			}
			log.Printf("connection closed, reason: %v\n", reason)

			for {
				time.Sleep(c.delay)

				err := c.dial()
				if err == nil {
					log.Println("reconnect success")
					break
				}

				log.Printf("reconnect failed, err: %v\n", err)
			}
		}
	}
}
