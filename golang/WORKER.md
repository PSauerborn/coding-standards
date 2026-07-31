---
title: Golang Worker Standards
description: Standards for writing background workers in Go.
scope:
- '*.go'
parent: golang/GENERAL.md
topics:
- golang
- worker
- rabbitmq
- message-broker
- amqp
examples:
- examples/WORKER/reference-worker.md
---

# Golang Worker Standards

## 1. Worker Guidelines

`[GO-WRK-001]` **MUST**: Workers must use RabbitMQ as the message broker.

`[GO-WRK-002]` **MUST**: Workers must use the `github.com/rabbitmq/amqp091-go` package to interact with RabbitMQ.

`[GO-WRK-003]` **MUST**: The main goroutine must be the one that initializes the RabbitMQ connection and listens for messages. Any other worker processes must be started as separate goroutines.

`[GO-WRK-004]` **MUST**: RabbitMQ interface layers must be thread-safe. Take particular care to avoid deadlocks when using channels.

`[GO-WRK-005]` **SHOULD**: RabbitMQ interface layers should be minimal. Messages should be passed to the worker goroutine(s) via a channel.

`[GO-WRK-006]` **SHOULD**: Avoid using the default exchange. Instead, use a custom exchange with a `topic` and declare the appropriate queues and bindings.

`[GO-WRK-007]` **SHOULD**: Exchanges and queues should be declared durable. This ensures that they are not lost in case of a failure.

`[GO-WRK-008]` **SHOULD**: The message prefetch count should be set to match the maximum concurrency of the worker. This ensures that the broker does not overwhelm the worker with messages.

`[GO-WRK-009]` **SHOULD**: Channels should be buffered to avoid deadlocks and limit concurrency. The buffer size should be equal to the maximum concurrency of the worker and prefetch count, which should always be configurable. Unbuffered channels can be used where applicable to control processes running on separate go routines, but this should be an intentional design choice.

`[GO-WRK-010]` **SHOULD**: Worker goroutines/processes should be able to signal the main goroutine that they have completed their tasks. This ensures that the RabbitMQ interface can control acknowledgment and requeueing.

`[GO-WRK-011]` **SHOULD**: Avoid auto-acknowledgment of messages. Instead, manually acknowledge messages after processing. This ensures that messages are not lost in case of a failure.

See `examples/WORKER/reference-worker.md` for a reference worker implementation.
