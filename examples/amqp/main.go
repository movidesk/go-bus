package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	base "github.com/movidesk/go-bus"
	"github.com/movidesk/go-bus/amqp"
)

func main() {
	bus, err := amqp.NewBus()
	if err != nil {
		log.Panic(err)
	}

	pub, err := bus.NewPublisher()
	if err != nil {
		log.Panic(err)
	}
	go publish(pub)

	sub, err := bus.NewSubscriber()
	if err != nil {
		log.Panic(err)
	}
	go subscribe(sub)

	close := make(chan os.Signal)
	signal.Notify(close, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-close
	bus.Close()
	bus.Wait()
}

func publish(pub amqp.Publisher) {
	for {
		time.Sleep(time.Second * 1)
		err, ok := pub.Publish(base.Message{})
		if err != nil {
			break
		}
		if ok {
			log.Println("publish confirmed")
		}
	}
}

func subscribe(sub amqp.Subscriber) {
	msgs, done, err := sub.Consume()
	if err != nil {
		return
	}
	for {
		select {
		case <-done:
			break
		case msg := <-msgs:
			log.Printf("%+v", msg)
		}
	}
}
