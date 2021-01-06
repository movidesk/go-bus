package amqp

import (
	"log"
	"time"

	"github.com/streadway/amqp"
)

func waitToBeTrue(check func() bool, d time.Duration) {
	end := time.Now().Add(d)
	for {
		if check() {
			return
		}

		if end.Before(time.Now()) {
			return
		}
	}
}

func waitForTimeout(fn func(), d time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		fn()
	}()
	select {
	case <-c:
		return false
	case <-time.After(d):
		return true
	}
}

func declareTopic(dsn, exchange, queue string) {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		log.Fatalf("cannot (re)dial: %v: %q", err, dsn)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("cannot create channel: %v", err)
	}

	if err := ch.ExchangeDeclare(exchange, "topic", false, true, false, false, nil); err != nil {
		log.Fatalf("cannot declare fanout exchange: %v", err)
	}

	if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		log.Fatalf("cannot declare queue: %v", err)
	}

	if err := ch.QueueBind(queue, "", exchange, false, nil); err != nil {
		log.Fatalf("cannot bind: %v", err)
	}
}

func deleteTopic(dsn, exchange, queue string) {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		log.Fatalf("cannot (re)dial: %v: %q", err, dsn)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("cannot create channel: %v", err)
	}

	if err := ch.QueueUnbind(queue, "", exchange, nil); err != nil {
		log.Fatalf("cannot unbind: %v", err)
	}

	if _, err := ch.QueueDelete(queue, false, false, false); err != nil {
		log.Fatalf("cannot delete queue: %v", err)
	}

	if err := ch.ExchangeDelete(exchange, false, false); err != nil {
		log.Fatalf("cannot delete exchange: %v", err)
	}
}
