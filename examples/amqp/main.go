package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/movidesk/go-bus/amqp"
	toxi "github.com/shopify/toxiproxy/client"
)

func main() {
	cli := toxi.NewClient("docker:8474")
	_, err := cli.Proxy("rabbit")
	if err != nil {
		_, err = cli.CreateProxy("rabbit", "docker:35672", "mq:5672")
	}

	bus, err := amqp.NewBus(amqp.SetBusDSN("amqp://guest:guest@docker:35672"))
	if err != nil {
		log.Panic(err)
	}

	events, err := bus.NewPublisher(amqp.SetPublisherExchange("events"))
	if err != nil {
		log.Panic(err)
	}
	go publish(events, "events", time.Second)

	eventsa, err := bus.NewSubscriber(amqp.SetSubscriberQueue("event_queue_a"))
	if err != nil {
		log.Panic(err)
	}
	go subscribe(eventsa, "eventsa", time.Second)

	eventsb, err := bus.NewSubscriber(amqp.SetSubscriberQueue("event_queue_b"))
	if err != nil {
		log.Panic(err)
	}
	go subscribe(eventsb, "eventsb", time.Second)

	commands, err := bus.NewPublisher(amqp.SetPublisherExchange("commands"))
	if err != nil {
		log.Panic(err)
	}
	go publish(commands, "commands", time.Second)

	workera, err := bus.NewSubscriber(amqp.SetSubscriberQueue("command_queue"))
	if err != nil {
		log.Panic(err)
	}
	go subscribe(workera, "workera", time.Second*2)

	workerb, err := bus.NewSubscriber(amqp.SetSubscriberQueue("command_queue"))
	if err != nil {
		log.Panic(err)
	}
	go subscribe(workerb, "workerb", time.Second*2)

	close := make(chan os.Signal)
	signal.Notify(close, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-close
	bus.Close()
	bus.Wait()
}

func publish(pub amqp.Publisher, name string, d time.Duration) {
	c := 0
	for {
		c++
		time.Sleep(d)
		err, ok := pub.Publish(&amqp.Message{
			Headers: map[string]interface{}{
				"counter": c,
			},
		})
		if err != nil {
			log.Printf("%s: %+v", name, err)
			continue
		}
		if ok {
			log.Printf("%s: publish confirmed\n", name)
		} else {
			log.Printf("%s: publish not confirmed\n", name)
		}
	}
}

func subscribe(sub amqp.Subscriber, name string, d time.Duration) {
	msgs, done, err := sub.Consume()
	if err != nil {
		return
	}
	for {
		time.Sleep(d)
		select {
		case <-done:
			break
		case msg := <-msgs:
			msg.Ack(false)
			log.Printf("%s: acked message %+v\n", name, msg)
		}
	}
}
