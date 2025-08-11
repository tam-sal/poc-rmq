package main

import (
	"log"
	handlers "processing/processingHandlers"

	amqp "github.com/rabbitmq/amqp091-go"
)

const numWorkers = 10
const queueName = "image_processing_queue"

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	jobs := make(chan amqp.Delivery, numWorkers)

	for i := 0; i < numWorkers; i++ {

		go handlers.Worker(i+1, ch, jobs)
	}

	// Start consuming messages from RabbitMQ and send them to the jobs channel
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack is false to manually acknowledge
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	log.Printf("Processing Service is running with %d workers and consuming messages from %s", numWorkers, queueName)

	for d := range msgs {
		jobs <- d
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
