package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	amqp "github.com/rabbitmq/amqp091-go"
)

func HandleUploadMessage(d amqp.Delivery, ch *amqp.Channel) {
	log.Printf("Received message body: %s", d.Body)

	var body struct {
		ImageID      string `json:"image_id"`
		OriginalPath string `json:"original_path"`
	}

	if err := json.Unmarshal(d.Body, &body); err != nil {
		log.Printf("Failed to unmarshal JSON: %v", err)
		d.Ack(false)
		return
	}

	log.Printf("Successfully parsed message. Image ID: %s", body.ImageID)

	fileExtension := filepath.Ext(body.OriginalPath)
	uuidFilename := fmt.Sprintf("%s%s", body.ImageID, fileExtension)
	inputPath := filepath.Join("storage", "images", "in", uuidFilename)

	if _, err := os.Stat(filepath.Dir(inputPath)); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(inputPath), os.ModePerm)
	}

	originalFile, err := os.Open(body.OriginalPath)
	if err != nil {
		log.Printf("Failed to open original file at %s: %v", body.OriginalPath, err)
		d.Ack(false)
		return
	}
	defer originalFile.Close()

	newFile, err := os.Create(inputPath)
	if err != nil {
		log.Printf("Failed to create new file at %s: %v", inputPath, err)
		d.Ack(false)
		return
	}
	defer newFile.Close()

	if _, err := io.Copy(newFile, originalFile); err != nil {
		log.Printf("Failed to copy file from %s to %s: %v", body.OriginalPath, inputPath, err)
		d.Ack(false)
		return
	}
	log.Printf("File copied from '%s' to '%s'", body.OriginalPath, inputPath)

	log.Println("Publishing message to image_processing_queue...")

	processingBody := map[string]string{
		"image_id":   body.ImageID,
		"input_path": inputPath,
	}

	processingBodyJSON, err := json.Marshal(processingBody)
	if err != nil {
		log.Printf("Failed to marshal JSON for processing message: %v", err)
		d.Ack(false)
		return
	}

	err = ch.Publish(
		"",
		"image_processing_queue",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        processingBodyJSON,
		})
	if err != nil {
		log.Printf("Failed to publish a message to the processing queue: %v", err)
		d.Ack(false)
		return
	}

	log.Printf("Published message to image_processing_queue: %s", processingBodyJSON)
	d.Ack(false)
	log.Println("Acknowledged message from upload queue.")
}
