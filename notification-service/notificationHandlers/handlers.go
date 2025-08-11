package notificationHandlers

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func HandleNotificationMessage(d amqp.Delivery) {
	log.Printf("Received notification message: %s", d.Body)

	// Acknowledge the message.
	d.Ack(false)

	// Deserialize the message to get the image details.
	var msg map[string]string
	err := json.Unmarshal(d.Body, &msg)
	if err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return
	}

	log.Printf("NOTIFICATION: Image with ID '%s' has been successfully processed and is ready at '%s'.",
		msg["image_id"], msg["grayscale_path"])
}
