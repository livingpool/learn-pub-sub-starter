package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril game client...")

	const rabbitConnString = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to rabbitmq: %v", err)
	}
	defer conn.Close()

	fmt.Println("Peril game client connected to rabbitmq!")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("could not use username: %v", err)
		return
	}

	_, queue, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, 1)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	gameState := gamelogic.NewGameState(username)
	for {
		words := gamelogic.GetInput()
		switch words[0] {
		case "spawn":
			err := gameState.CommandSpawn(words)
			if err != nil {
				fmt.Println(err)
			}
		case "move":
			_, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Println(err)
			}
			// TODO: publish the move
		case "status":
			gameState.CommandStatus()
		case "spam":
			// TODO: publish n malicious logs
			fmt.Println("Spamming not allowed yet!")
		case "help":
			gamelogic.PrintClientHelp()
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("unknown command:", words[0])
		}
	}
}
