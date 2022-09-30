package main

import (
	api "shoulder/api"
	gen "shoulder/api/gen"

	"log"

	middleware "github.com/deepmap/oapi-codegen/pkg/gin-middleware"
	"github.com/gin-gonic/gin"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	swagger, err := gen.GetSwagger()
	if err != nil {
		return
	}

	r := gin.Default()

	r.Use(middleware.OapiRequestValidator(swagger))

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Panicf("%s: %s", "Failed to connect to RabbitMQ", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Panicf("%s: %s", "Failed to open a channel", err)
	}

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		log.Panicf("%s: %s", "Failed to declare a queue", err)
	}

	pingAPI := api.PingAPI{Channel: ch, Queue: q}

	gen.RegisterHandlers(r, pingAPI)
	if err != nil {
		return
	}
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
