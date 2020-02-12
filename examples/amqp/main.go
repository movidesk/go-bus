package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/movidesk/go-bus/amqp"
)

func main() {
	bus, err := amqp.NewBus()
	if err != nil {
		log.Panic(err)
	}

	events, err := bus.NewPublisher(amqp.SetPublisherExchange("events"))
	if err != nil {
		log.Panic(err)
	}
	go publish(events, "events")

	eventsa, err := bus.NewSubscriber(amqp.SetSubscriberQueue("event_queue_a"))
	if err != nil {
		log.Panic(err)
	}
	go subscribe(eventsa, "eventsa")

	eventsb, err := bus.NewSubscriber(amqp.SetSubscriberQueue("event_queue_b"))
	if err != nil {
		log.Panic(err)
	}
	go subscribe(eventsb, "eventsb")

	commands, err := bus.NewPublisher(amqp.SetPublisherExchange("commands"))
	if err != nil {
		log.Panic(err)
	}
	go publish(commands, "commands")

	workera, err := bus.NewSubscriber(amqp.SetSubscriberQueue("command_queue"))
	if err != nil {
		log.Panic(err)
	}
	go subscribe(workera, "workera")

	workerb, err := bus.NewSubscriber(amqp.SetSubscriberQueue("command_queue"))
	if err != nil {
		log.Panic(err)
	}
	go subscribe(workerb, "workerb")

	close := make(chan os.Signal)
	signal.Notify(close, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-close
	bus.Close()
	bus.Wait()
}

func publish(pub amqp.Publisher, name string) {
	c := 0
	for {
		c++
		time.Sleep(time.Second * 1)
		err, ok := pub.Publish(&amqp.Message{
			Headers: map[string]interface{}{
				"counter": c,
			},
		})
		if err != nil {
			break
		}
		if ok {
			log.Printf("%s: publish confirmed by\n", name)
		}
	}
}

func subscribe(sub amqp.Subscriber, name string) {
	msgs, done, err := sub.Consume()
	if err != nil {
		return
	}
	for {
		select {
		case <-done:
			break
		case msg := <-msgs:
			msg.Ack(false)
			log.Printf("%s: acked message %+v\n", name, msg)
		}
	}
}
