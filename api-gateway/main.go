package main

import (
	"api-gateway/handlers"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// RabbitMQ connection
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Open a channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// RMQ channel to the handlers
	gatewayHandlers := handlers.NewGatewayHandlers(ch)

	router := mux.NewRouter()
	router.HandleFunc("/upload", gatewayHandlers.HandleUpload).Methods("POST")

	fmt.Println("API Gateway is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
