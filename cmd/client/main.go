package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"strconv"
	"time"
)

func main() {
	fmt.Println("Starting Peril client...")

	const rabbitURL = "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Error Dialing to RabbitMQ %v", err)
	}
	defer connection.Close()
	fmt.Println("Client's connection to RabbitMQ successful")

	channel, err := connection.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Could not get username: %v", err)
	}

	gameState := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gameState),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gameState, channel),
	)
	if err != nil {
		log.Fatalf("could not subscribe to army moves: %v", err)
	}

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
		handlerWar(gameState, channel),
	)
	if err != nil {
		log.Fatalf("could not subscribe to war declarations: %v", err)
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			err := gameState.CommandSpawn(words)
			if err != nil {
				fmt.Printf("Error spawning game: %s", err)
				continue
			}
		case "move":
			armyMove, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Printf("Error moving: %s", err)
				continue
			}
			fmt.Printf("move %s %d", armyMove.ToLocation, len(armyMove.Units))
			err = pubsub.PublishJSON(channel, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, armyMove)
			if err != nil {
				fmt.Printf("Error moving: %s", err)
				continue
			}
			fmt.Printf("move %s %d was published successfully", armyMove.ToLocation, len(armyMove.Units))
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			s, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Printf("error parsing words: %s", err)
				continue
			}
			for _ = range s {
				maliciousLog := gamelogic.GetMaliciousLog()
				err := pubsub.PublishGob(channel, routing.ExchangePerilTopic, routing.GameLogSlug+"."+username, routing.GameLog{
					CurrentTime: time.Now(),
					Message:     maliciousLog,
					Username:    username,
				})
				if err != nil {
					fmt.Printf("Error publishing smap: %s", err)
					continue
				}
			}
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Printf("Command unknown\n")
		}
	}
}
