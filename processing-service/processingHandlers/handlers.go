package handlers

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"

	amqp "github.com/rabbitmq/amqp091-go"
)

const publishQueueName = "image_ready_queue"

type ReadyMessage struct {
	ImageID       string `json:"image_id"`
	GrayscalePath string `json:"grayscale_path"`
}

func Worker(id int, ch *amqp.Channel, jobs <-chan amqp.Delivery) {
	log.Printf("Worker %d started", id)
	for d := range jobs {
		HandleProcessingMessage(d, ch)
	}
}

func HandleProcessingMessage(d amqp.Delivery, ch *amqp.Channel) {
	log.Printf("Received message for processing: %s", d.Body)

	d.Ack(false)

	var msg map[string]string
	err := json.Unmarshal(d.Body, &msg)
	if err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return
	}

	imagePath := msg["input_path"]
	imageID := msg["image_id"]

	absoluteInputPath := filepath.Join("/app", imagePath)
	file, err := os.Open(absoluteInputPath)
	if err != nil {
		log.Printf("Error opening file %s: %v", absoluteInputPath, err)
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Printf("Error decoding image: %v", err)
		return
	}

	// Convert the image to grayscale
	bounds := img.Bounds()
	grayImage := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			grayColor := color.GrayModel.Convert(originalColor)
			grayImage.Set(x, y, grayColor)
		}
	}

	// Save the grayscale image to the shared storage
	outputFileName := fmt.Sprintf("%s_grayscale.png", imageID)
	outputFilePath := filepath.Join("/app", "storage", "images", "out", outputFileName)
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		log.Printf("Error creating output file: %v", err)
		return
	}
	defer outputFile.Close()

	if err = png.Encode(outputFile, grayImage); err != nil {
		log.Printf("Error encoding image to PNG: %v", err)
		return
	}
	log.Printf("Grayscale image saved at: %s", outputFilePath)

	// Publish a new message to the "image_ready" queue
	readyMsg := ReadyMessage{
		ImageID:       imageID,
		GrayscalePath: outputFilePath,
	}

	readyMsgBody, err := json.Marshal(readyMsg)
	if err != nil {
		log.Printf("Error marshalling ready message: %v", err)
		return
	}

	err = ch.Publish(
		"",               // exchange
		publishQueueName, // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        readyMsgBody,
		})
	if err != nil {
		log.Printf("Failed to publish ready message: %v", err)
	}

	log.Printf("Published 'image ready' message for ID: %s", imageID)
}
