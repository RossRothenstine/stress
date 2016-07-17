package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/streadway/amqp"
	"gitlab.com/Butabah/stress"
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
)

var (
	msgsPerSecond int
	amqpURLString string
	exchange      string
	body          string
)

func init() {
	flag.IntVar(&msgsPerSecond, "amount", 10, "Number of messages to publish per second")
	flag.StringVar(&amqpURLString, "url", "amqp:///", "The AMQP connection string")
	flag.StringVar(&exchange, "exchange", "amq.fanout", "The AMQP connection string")
	flag.StringVar(&body, "body", "{}", "The body to publish")
}

func main() {
	os.Exit(Main())
}

func Main() int {
	flag.Parse()
	if msgsPerSecond <= 0 {
		fmt.Printf("Messages per second cannot be <= 0")
		return 1
	}

	conn, err := amqp.Dial(amqpURLString)
	if err != nil {
		fmt.Printf("dial: %v\n", err)
		return 1
	}
	closed := make(chan *amqp.Error)
	conn.NotifyClose(closed)

	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("channel: %v\n", err)
		return 1
	}

	stresser := &stress.Stresser{
		Limit: rate.NewLimiter(rate.Limit(msgsPerSecond), msgsPerSecond),
		New: func() stress.Worker {
			return &AMQPPublisherWorker{
				Body:     body,
				Channel:  ch,
				Exchange: exchange,
			}
		},
	}
	go stresser.Start()
	defer stresser.Stop()

	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Kill, os.Interrupt)

	select {
	case <-interrupt:
		return 0
	case err := <-closed:
		fmt.Printf("amqp: ", err)
		return 1
	}

}

type AMQPPublisherWorker struct {
	Body     string
	Exchange string
	Channel  *amqp.Channel
}

func (worker *AMQPPublisherWorker) Work(context.Context) {
	err := worker.Channel.Publish(
		worker.Exchange,
		"",
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(worker.Body),
		},
	)
	if err != nil {
		fmt.Printf("err occurred publishing: %v\n", err)
	}
}
