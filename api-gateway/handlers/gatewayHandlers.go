package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const queueName = "image_upload_queue"

type GatewayHandlers struct {
	ch *amqp.Channel
}

func NewGatewayHandlers(ch *amqp.Channel) *GatewayHandlers {
	return &GatewayHandlers{ch: ch}
}

func (h *GatewayHandlers) HandleUpload(w http.ResponseWriter, r *http.Request) {
	// 5MB max file size
	r.ParseMultipartForm(5 << 20)

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Invalid file upload", http.StatusBadRequest)
		log.Printf("Error retrieving the file: %v", err)
		return
	}
	defer file.Close()

	originalFilename := fileHeader.Filename
	fileUUID := uuid.New().String()
	originalPath := filepath.Join("storage", "images", "origin", originalFilename)

	if _, err := os.Stat(filepath.Dir(originalPath)); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(originalPath), os.ModePerm)
	}

	dst, err := os.Create(originalPath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		log.Printf("Error creating destination file: %v", err)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to copy file", http.StatusInternalServerError)
		log.Printf("Error copying file: %v", err)
		return
	}
	log.Printf("File '%s' uploaded and saved at: %s", fileHeader.Filename, originalPath)

	message := map[string]string{
		"image_id":      fileUUID,
		"original_path": originalPath,
	}
	body, _ := json.Marshal(message)

	err = h.ch.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		log.Printf("Failed to publish message: %v", err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(fmt.Sprintf("Request received. Image ID: %s", fileUUID)))
}
