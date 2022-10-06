package internal

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func GetChannel(shoulderConfig ShoulderConfig, exchangeType string) (*amqp.Channel, amqp.Queue) {
	conn, err := amqp.Dial(shoulderConfig.AmqpConnString)
	if err != nil {
		log.Panicf("%s: %s", "Failed to connect to RabbitMQ", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Panicf("%s: %s", "Failed to open a channel", err)
	}

	err = ch.ExchangeDeclare(
		shoulderConfig.ExchangeName,
		exchangeType,
		true, false, false, false, nil,
	)
	if err != nil {
		log.Panicf("%s: %s", "Failed to declare an exchange", err)
	}

	q, err := ch.QueueDeclare(
		"", false, false, true, false, nil,
	)
	if err != nil {
		log.Panicf("%s: %s", "Failed to declare a queue", err)
	}

	err = ch.QueueBind(
		q.Name, "", shoulderConfig.ExchangeName, false, nil,
	)
	if err != nil {
		log.Panicf("%s: %s", "Failed to bind a queue", err)
	}
	return ch, q
}
