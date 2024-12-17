package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func main() {
	fmt.Println("Starting Peril server...")

	const rabbitURL = "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Error Dialing to RabbitMQ %v", err)
	}
	defer connection.Close()
	fmt.Println("Connection to RabbitMQ successful")

	channel, err := connection.Channel()
	if err != nil {
		log.Fatalf("Error creating new channel %v", err)
	}

	_, queue, err := pubsub.DeclareAndBind(connection, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.SimpleQueueDurable)
	if err != nil {
		log.Fatalf("Could not bind queue to pause: %v", err)
	}
	fmt.Printf("Queue %v declared and bound\n", queue.Name)

	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "pause":
			log.Println("sending a pause message")
			err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			if err != nil {
				log.Printf("Error publishing to channel: %v", err)
			}
		case "resume":
			log.Println("sending a resume message")
			err = pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			if err != nil {
				log.Printf("Error publishing to channel: %v", err)
			}
		case "quit":
			log.Println("Exiting server")
			return
		default:
			log.Println("Don't understand the command")
		}
	}
}
