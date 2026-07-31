# [GO-WRK-001] Reference Worker Implementation

Statements: `[GO-WRK-001]` `[GO-WRK-002]` `[GO-WRK-003]` `[GO-WRK-004]` `[GO-WRK-005]` `[GO-WRK-006]` `[GO-WRK-007]` `[GO-WRK-008]` `[GO-WRK-009]` `[GO-WRK-010]` `[GO-WRK-011]`

The following process illustrates how to implement a worker process using the above guidelines:

```go
// GOOD
// filename: main.go

import (
    amqp "github.com/rabbitmq/amqp091-go"
)

const (
    MaxConcurrency = 15
)

type EventResult struct {
    Metadata interface{}
    Requeue  bool
}

type CustomEventType struct {
    Foo string `json:"foo"`
    Metadata   interface{}
}

type RabbitMQBroker struct {
    conn *amqp.Connection
    ch   *amqp.Channel
}

// NewRabbitMQBroker creates a new RabbitMQ broker
func NewRabbitMQBroker() (*RabbitMQBroker, error) {
    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        return nil, err
    }
    ch, err := conn.Channel()
    if err != nil {
        return nil, err
    }
    return &RabbitMQBroker{conn: conn, ch: ch}, nil
}

// Close the connection to RabbitMQ
func (b *RabbitMQBroker) Close() error {
    if err := b.ch.Close(); err != nil {
        return err
    }
    return b.conn.Close()
}

// GOOD: ack messages manually
// AckMessage manually acknowledges a message
func (b *RabbitMQBroker) AckMessage(delivery *amqp.Delivery) error {
    if err := delivery.Ack(false); err != nil {
        return err
    }
    log.WithFields(log.Fields{
        "delivery_tag": delivery.DeliveryTag,
    }).Info("acked message")
    return nil
}

// GOOD: events and results are communicated via channels
// Listen to messages from RabbitMQ and pass them to the worker
func (b *RabbitMQBroker) Listen(events chan CustomEventType, results chan EventResult) error {
    // GOOD: avoid using the default exchange. declare a dedicated exchange
    _, err := b.ch.ExchangeDeclare(
        "user.events",
        "topic",
        true,
		false,
		false,
		false,
		nil,
    )
    if err != nil {
        return err
    }

    // GOOD: declare a dedicated queue
    queue, err := b.ch.QueueDeclare(
        "example.process-events",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

    // GOOD: bind any routing keys
    if err := b.ch.QueueBind(
        queue.Name,
        "user.created",
        "user.events",
        false,
        nil,
    ); err != nil {
        return err
    }

    deliveries, err := b.ch.Consume(
        queue.Name,
        "",
        false,
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        return err
    }

    // GOOD: run ack on separate goroutine
    go func() {
        for result := range results {
            delivery := result.Metadata.(amqp.Delivery)
            if result.Requeue {
                if err := b.ch.Nack(delivery.DeliveryTag, false, true); err != nil {
                    log.WithError(err).Error("failed to nack message")
                }
            } else {
                if err := b.AckMessage(&delivery); err != nil {
                    log.WithError(err).Error("failed to ack message")
                }
            }
        }
    }()

    for delivery := range deliveries {
        var e CustomEventType
        if err := json.Unmarshal(delivery.Body, &e); err != nil {
            return err
        }
        e.Metadata = delivery

        events <- e
    }
    return nil
}

func main() {
    // GOOD: use channels to communicate between goroutines
    // GOOD: use buffered channels to limit concurrency within a process and prevent
    // deadlocks
    events := make(chan CustomEventType, MaxConcurrency)
    results := make(chan EventResult)

    broker, err := NewRabbitMQBroker()
    if err != nil {
        log.Fatal(err)
    }

    // GOOD: worker goroutine processes events
    go func() {
        for event := range events {
            // GOOD: implement structured logging
            log.WithFields(log.Fields{
                "event": event,
            }).Info("processing event")

            // GOOD: communicate task completion to main goroutine
            results <- EventResult{
                Metadata: event,
                Requeue:  false,
            }
        }
    }()

    // GOOD: main goroutine listens for messages
    if err := broker.Listen(events, results); err != nil {
        log.WithError(err).Error("failed to listen for messages")
        panic(err)
    }
}
```
