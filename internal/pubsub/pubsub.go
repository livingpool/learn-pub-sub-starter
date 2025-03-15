package pubsub

import (
	"context"
	"encoding/json"
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return nil
	}

	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	}

	ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	queue := amqp.Queue{}
	ch, err := conn.Channel()
	if err != nil {
		return nil, queue, err
	}

	if simpleQueueType == 0 { // durable
		q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			return nil, q, err
		}
		queue = q
	} else if simpleQueueType == 1 { // transient
		q, err := ch.QueueDeclare(queueName, false, true, true, false, nil)
		if err != nil {
			return nil, q, err
		}
		queue = q
	} else {
		return nil, queue, errors.New("unrecognized queue type")
	}

	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, queue, err
	}

	return ch, queue, nil
}
