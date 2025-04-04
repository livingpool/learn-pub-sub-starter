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

	publishChan, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("could not use username: %v", err)
		return
	}

	gs := gamelogic.NewGameState(username)

	// bind to the pause routing key
	err = pubsub.SubscribeJSON(conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gs),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	// bind to the army_moves.* routing key
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gs, publishChan),
	)
	if err != nil {
		log.Fatalf("could not subscribe to army move: %v", err)
	}

	// to consumes all the war messages that the "move" handler publishes, no matter the username in the routing key
	// bind to the war.* routing key
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
		handlerWar(gs, publishChan),
	)
	if err != nil {
		log.Fatalf("could not subscribe to army move: %v", err)
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		switch words[0] {
		case "spawn":
			err := gs.CommandSpawn(words)
			if err != nil {
				fmt.Println(err)
			}
		case "move":
			mv, err := gs.CommandMove(words)
			if err != nil {
				fmt.Println(err)
			}
			pubsub.PublishJSON(publishChan, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, mv)
			fmt.Printf("Moved %v units to %s\n", len(mv.Units), mv.ToLocation)
		case "status":
			gs.CommandStatus()
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
