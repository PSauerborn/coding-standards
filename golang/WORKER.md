---
title: Golang Worker Standards
description: Standards for writing background workers in Go.
scope: '*.go'
parent: golang/GENERAL.md
topics:
- golang
- worker
- rabbitmq
- message-broker
- amqp
---

## 1. Meta Rules

You are a Senior Software Engineer acting as an autonomous coding agent.

1.  **Strict Adherence**: You MUST follow all **MUST** rules below.
2.  **Pattern Matching**: When writing code, check the "Example" sections. If you are tempted to write code that looks like a "BAD" example, STOP and refactor to match the "GOOD" example.
3.  **Explanation**: If you deviate from a **SHOULD** rule, you must explicitly state why in your reasoning trace.

If a user request contradicts a **SHOULD** statement, follow the user request. If it contradicts a **MUST** statement, ask for confirmation.

## 2. Worker Guidelines

**MUST**: Workers must use RabbitMQ as the message broker.

**MUST**: Workers must use the `github.com/rabbitmq/amqp091-go` package to interact with RabbitMQ.

**MUST**: The main goroutine must be the one that initializes the RabbitMQ connection and listens for messages. Any other worker processes must be started as separate goroutines.

**MUST**: RabbitMQ interface layers must be thread-safe. Take particular care to avoid deadlocks when using channels.

**SHOULD**: RabbitMQ interface layers should be minimal. Messages should be passed to the worker goroutine(s) via a channel.

**SHOULD**: Avoid using the default exchange. Instead, use a custom exchange with a `topic` and declare the appropriate queues and bindings.

**SHOULD**: Exchanges and queues should be declared durable. This ensures that they are not lost in case of a failure.

**SHOULD**: The message prefetch count should be set to match the maximum concurrency of the worker. This ensures that the broker does not overwhelm the worker with messages.

**SHOULD**: Channels should be buffered to avoid deadlocks. The buffer size should be equal to the maximum concurrency of the worker and prefetch count.

**SHOULD**: Worker goroutines/processes should be able to signal the main goroutine that they have completed their tasks. This ensures that the RabbitMQ interface can control acknowledgment and requeueing.

**SHOULD**: Avoid auto-acknowledgment of messages. Instead, manually acknowledge messages after processing. This ensures that messages are not lost in case of a failure.

### 3. Example

The following process illustrates how to implement a worker process using the above guidelines:

```go
// GOOD
// filename: main.go

import (
    amqp "github.com/rabbitmq/amqp091-go"
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
    events := make(chan CustomEventType)
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
