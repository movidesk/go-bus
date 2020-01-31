package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	base "github.com/movidesk/go-bus"
	"github.com/movidesk/go-bus/proc"
)

var broker = make(chan base.Message, 10)

func main() {
	bus, err := proc.NewBus(
		proc.SetIn(broker),
		proc.SetOut(broker),
	)
	if err != nil {
		log.Fatal("Unable to create in process bus")
	}

	pub, err := bus.NewPublisher()
	if err != nil {
		log.Fatal("unable to create publisher")
	}

	sub, err := bus.NewSubscriber()
	if err != nil {
		log.Fatal("unable to create subscriber")
	}

	go func() {
		out, closer, err := sub.Consume()
		defer sub.Close()

		if err != nil {
			log.Fatal("unable to subscribe")
		}

		done := false
		for !done {
			time.Sleep(time.Second * 2)
			select {
			case <-closer:
				done = true
			case msg := <-out:
				log.Printf("(%d): %+v\n", len(out), msg)
			}
		}

		for len(out) > 0 {
			msg := <-out
			log.Printf("(%d): %+v\n", len(out), msg)
		}

	}()

	go func() {
		for {
			time.Sleep(time.Second)
			err, ok := pub.Publish(base.Message{})
			if err != nil {
				log.Fatal("unable to publish message")
			}
			if ok {
				log.Println("publish has been confirmed")
			} else {
				log.Println("publish has not been confirmed")
			}
		}
	}()

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("quiting server")
		log.Println("closing bus")
		bus.Close()

	}()

	bus.Wait()
	log.Println("server quited")
	log.Println("bus closed")
}
