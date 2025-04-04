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
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to rabbitmq: %v", err)
	}
	defer conn.Close()

	fmt.Println("Peril game server connected to rabbitmq!")

	publishChan, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.SimpleQueueDurable,
		handlerGameLogs(),
	)
	if err != nil {
		log.Fatalf("could not consume game_logs: %v", err)
	}

	gamelogic.PrintServerHelp()
	for {
		words := gamelogic.GetInput()
		if words == nil {
			continue
		}

		switch words[0] {
		case "pause":
			pubsub.PublishJSON(publishChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			fmt.Println("pause message sent!")
		case "resume":
			pubsub.PublishJSON(publishChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			fmt.Println("resume message sent!")
		case "help":
			gamelogic.PrintServerHelp()
		case "quit":
			fmt.Println("exiting...")
			return
		default:
			fmt.Println("unknown command:", words[0])
		}
	}
}
