package main

import (
	"log"
	"notification-service/notificationHandlers"

	amqp "github.com/rabbitmq/amqp091-go"
)

// The name of the queue we will be consuming from.
const queueName = "image_ready_queue"

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

	log.Printf("Notification Service is connected and consuming messages from %s", queueName)
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			notificationHandlers.HandleNotificationMessage(d)
		}
	}()

	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
