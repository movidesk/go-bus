package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/movidesk/go-bus/amqp"
)

func main() {
	done := make(chan struct{})

	go proxy(":25672", "localhost:5672")(done)

	sess, err := amqp.NewSession(
		amqp.SetDsn("amqp://admin:admin@localhost:25672"),
	)
	if err != nil {
		log.Fatal("amqp.NewSession")
	}

	for !sess.IsConnected() {
		time.Sleep(time.Second)
		log.Println("not connected!")
	}

	time.Sleep(time.Second * 10)
	done <- struct{}{}
	log.Println("proxy disconnect")

	for sess.IsConnected() {
		time.Sleep(time.Second)
		log.Println("connected!")
	}

	go proxy(":25672", "localhost:5672")(done)

	for !sess.IsConnected() {
		time.Sleep(time.Second)
		log.Println("not connected!")
	}

	for sess.IsConnected() {
		time.Sleep(time.Second)
		log.Println("connected!")
	}
}

func proxy(from, to string) func(<-chan struct{}) {
	return func(done <-chan struct{}) {
		incoming, err := net.Listen("tcp", from)
		if err != nil {
			log.Fatalf("could not start server on %s: %v", from, err)
		}
		defer incoming.Close()
		fmt.Printf("server running on %s\n", from)

		client, err := incoming.Accept()
		if err != nil {
			log.Fatal("could not accept client connection", err)
		}
		defer client.Close()
		fmt.Printf("client '%v' connected!\n", client.RemoteAddr())

		target, err := net.Dial("tcp", to)
		if err != nil {
			log.Fatal("could not connect to target", err)
		}
		defer target.Close()
		fmt.Printf("connection to server %v established!\n", target.RemoteAddr())

		go func() { io.Copy(target, client) }()
		go func() { io.Copy(client, target) }()

		<-done
	}
}
