package amqp

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type ChannelOptionsFn func(*ChannelOptions)

type ChannelOptions struct {
	delay time.Duration

	prefetchCount int
	prefetchSize  int

	wg   *sync.WaitGroup
	done <-chan struct{}
}

func SetChannelDelay(delay time.Duration) ChannelOptionsFn {
	return func(o *ChannelOptions) {
		o.delay = delay
	}
}

func SetChannelPrefetchCount(count int) ChannelOptionsFn {
	return func(o *ChannelOptions) {
		o.prefetchCount = count
	}
}

func SetChannelPrefetchSize(size int) ChannelOptionsFn {
	return func(o *ChannelOptions) {
		o.prefetchSize = size
	}
}

func SetChannelWaitGroup(wg *sync.WaitGroup) ChannelOptionsFn {
	return func(o *ChannelOptions) {
		o.wg = wg
	}
}

func SetChannelDone(done <-chan struct{}) ChannelOptionsFn {
	return func(o *ChannelOptions) {
		o.done = done
	}
}

type Channel struct {
	*amqp.Channel
	*ChannelOptions
	conn   *Connection
	closed int32
}

func NewChannel(conn *Connection, fns ...ChannelOptionsFn) (*Channel, error) {
	o := &ChannelOptions{
		wg:    &sync.WaitGroup{},
		delay: time.Second,
	}
	SetChannelPrefetchCount(1)(o)
	SetChannelPrefetchSize(0)(o)
	for _, fn := range fns {
		fn(o)
	}

	chnn := &Channel{
		ChannelOptions: o,
		conn:           conn,
	}

	err := chnn.channel()
	if err != nil {
		return nil, err
	}

	chnn.wg.Add(1)
	go chnn.loop()

	return chnn, nil
}

func (c *Channel) IsClosed() bool {
	return (atomic.LoadInt32(&c.closed) == 1)
}

func (c *Channel) Close() error {
	if c.IsClosed() {
		return amqp.ErrClosed
	}
	atomic.StoreInt32(&c.closed, 1)
	return c.Channel.Close()
}

func (c *Channel) channel() error {
	chnn, err := c.conn.Channel()
	if err != nil {
		return errors.Wrap(err, "Could not conn.Channel()")
	}
	err = chnn.Qos(c.prefetchCount, c.prefetchSize, false)
	if err != nil {
		return err
	}

	c.Channel = chnn
	return nil
}

func (c *Channel) loop() {
	defer c.wg.Done()
	running := true

out:
	for running {
		select {
		case <-c.done:
			c.Close()

		case reason, ok := <-c.Channel.NotifyClose(make(chan *amqp.Error)):
			if !ok {
				log.Println("channel closed")
				atomic.StoreInt32(&c.closed, 1)
				running = false
				break out
			}
			log.Printf("channel closed, reason: %v\n", reason)
			atomic.StoreInt32(&c.closed, 1)

			for {
				time.Sleep(c.delay)

				err := c.channel()
				if err == nil {
					log.Println("channel reconnect success")
					atomic.StoreInt32(&c.closed, 0)
					break
				}

				log.Printf("channel reconnect failed, err: %v\n", err)
			}
		}
	}
}
