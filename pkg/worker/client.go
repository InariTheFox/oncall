package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitWorker struct {
	client       *amqp.Channel
	closeChannel chan *amqp.Error
	context      context.Context
	handlers     map[JobType]JobHandler
	pollInterval time.Duration
	queue        amqp.Queue
	queueName    string
}

var _ Worker = &RabbitWorker{}

func NewRabbitWorker(host, username, password, vhost string, port int, queueName, exchangeName string, pollInterval time.Duration) (*RabbitWorker, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d%s", username, password, host, port, vhost)
	c, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("Cannot dial %s, %w\n", url, err)
	}

	fmt.Printf("Connected to amqp://%s:%d%s\n", host, port, vhost)

	ch, err := c.Channel()
	if err != nil {
		return nil, fmt.Errorf("Cannot open channel. %w\n", err)
	}

	// Setup fair dispatching scheme
	ch.Qos(1, 0, false)

	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to declare queue %s, %w\n", queueName, err)
	}

	fmt.Printf("Queue '%s' created\n", queueName)

	err = ch.ExchangeDeclare(
		exchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to declare exchange %s, %w\n", exchangeName, err)
	}

	ch.QueueBind(
		queueName,
		"*",
		exchangeName,
		true,
		nil,
	)

	nc := c.NotifyClose(make(chan *amqp.Error, 1))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		select {
		case err, ok := <-nc:
			if !ok {
				nc = nil
			} else {
				fmt.Errorf("Channel closed %w\n", err)
			}
		}
	}()

	return &RabbitWorker{
		client:       ch,
		closeChannel: nc,
		context:      ctx,
		queue:        q,
		queueName:    queueName,
	}, nil
}

func (w *RabbitWorker) RegisterHandler(t JobType, h JobHandler, args any) {
	if w.handlers == nil {
		w.handlers = make(map[JobType]JobHandler)
	}

	w.handlers[t] = h
}

func (w *RabbitWorker) Run(ctx context.Context) error {

	var forever chan struct{}

	go func() {
		fmt.Printf("Waiting for messages on queue '%s'\n", w.queueName)

		msg, err := w.client.Consume(
			w.queueName,
			"worker",
			false,
			false,
			false,
			true,
			nil,
		)
		if err != nil {
			fmt.Errorf("Failed to consume message %w\n", err)
		}

		for d := range msg {
			for t, h := range w.handlers {
				if d.RoutingKey == string(t) {
					job := &Job{}

					if err = json.Unmarshal(d.Body, job); err != nil {
						fmt.Errorf("Unable to deserialize message %s\n", d.Body)
						return
					}

					d.Ack(false)
					h(ctx, job)
				}
			}
		}

		<-ctx.Done()
		fmt.Println("Shutting down consumer")
	}()

	<-forever

	fmt.Println("Closed listener")

	return nil
}

func (w *RabbitWorker) Stop(ctx context.Context) {
	w.client.Close()
}
