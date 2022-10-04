package main

import (
	"encoding/json"
	"os"
	api "shoulder/api"
	gen "shoulder/api/gen"
	database "shoulder/db"

	"log"

	middleware "github.com/deepmap/oapi-codegen/pkg/gin-middleware"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ShoulderConfig struct {
	amqpConnString     string
	postgresConnString string
}

func getConfig() ShoulderConfig {
	amqp := os.Getenv("AMQP_CONN_STRING")
	if amqp == "" {
		panic("No value for amqp connection string")
	}

	postgres := os.Getenv("POSTGRES_CONN_STRING")
	if postgres == "" {
		panic("No value for postgres connection string")
	}

	return ShoulderConfig{amqpConnString: amqp, postgresConnString: postgres}
}

func main() {
	swagger, err := gen.GetSwagger()
	if err != nil {
		return
	}

	r := gin.Default()

	r.Use(middleware.OapiRequestValidator(swagger))

	shoulderConfig := getConfig()

	conn, err := amqp.Dial(shoulderConfig.amqpConnString)
	if err != nil {
		log.Panicf("%s: %s", "Failed to connect to RabbitMQ", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Panicf("%s: %s", "Failed to open a channel", err)
	}

	exchangeName := "ping"

	err = ch.ExchangeDeclare(
		exchangeName,
		"fanout",
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
		q.Name, "", exchangeName, false, nil,
	)
	if err != nil {
		log.Panicf("%s: %s", "Failed to bind a queue", err)
	}

	db, err := gorm.Open(postgres.Open(shoulderConfig.postgresConnString), &gorm.Config{})
	if err != nil {
		log.Panicf("%s: %s", "Failed to connect to database", err)
	}

	database.InitDb(db)

	pingAPI := api.PingAPI{Channel: ch, ExchangeName: exchangeName, State: make(map[int]string)}
	project(db, pingAPI.State)

	msgs, err := pingAPI.Channel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Println("Could not consume")
	}

	go listen(msgs, pingAPI.State, db)

	gen.RegisterHandlers(r, pingAPI)
	if err != nil {
		return
	}
	r.Run()
}

// get all records from the db and put them in memory LOL
func project(db *gorm.DB, state map[int]string) {
	var accretions []database.Accretion
	db.Find(&accretions)
	for _, value := range accretions {
		state[value.Key] = value.Value
	}
}

// listen on the message channel we get and put messages in state and persist
func listen(msgs <-chan amqp.Delivery, state map[int]string, db *gorm.DB) {
	var message gen.CommandContent
	for msg := range msgs {
		err := json.Unmarshal(msg.Body, &message)
		if err != nil {
			log.Println("Could not unmarshal message")
		}
		state[message.Key] = state[message.Key] + message.Value

		accretions := []database.Accretion{}
		result := db.Where("key = ?", message.Key).Find(&accretions)

		accretion := database.Accretion{Key: message.Key, Value: message.Value}
		if result.RowsAffected == 0 {
			db.Create(&accretion)
		} else {
			accretion.Value = accretions[0].Value + message.Value
			db.Save(&accretion)
		}
	}
}
