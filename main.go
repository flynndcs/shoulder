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

const (
	EXCHANGE_TYPE = "fanout"
)

type ShoulderConfig struct {
	AmqpConnString     string
	PostgresConnString string
	ExchangeName       string
}

func main() {
	swagger, err := gen.GetSwagger()
	if err != nil {
		log.Panicf("%s: %s", "Failed to get swagger spec", err)
	}

	r := gin.Default()

	r.Use(middleware.OapiRequestValidator(swagger))

	shoulderConfig := getConfig()

	conn, err := amqp.Dial(shoulderConfig.AmqpConnString)
	if err != nil {
		log.Panicf("%s: %s", "Failed to connect to RabbitMQ", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Panicf("%s: %s", "Failed to open a channel", err)
	}

	err = ch.ExchangeDeclare(
		shoulderConfig.ExchangeName,
		EXCHANGE_TYPE,
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

	db, err := gorm.Open(postgres.Open(shoulderConfig.PostgresConnString), &gorm.Config{})
	if err != nil {
		log.Panicf("%s: %s", "Failed to connect to database", err)
	}

	database.InitDb(db)

	shoulderAPI := api.ShoulderAPI{Channel: ch, ExchangeName: shoulderConfig.ExchangeName, State: make(map[int]string)}
	project(db, shoulderAPI.State)

	msgs, err := shoulderAPI.Channel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Panicf("%s: %s", "Could not consume message", err)
	}

	go listen(msgs, shoulderAPI.State, db)

	gen.RegisterHandlers(r, shoulderAPI)
	if err != nil {
		log.Panicf("%s: %s", "Could not create API", err)
	}
	err = r.Run()
	if err != nil {
		log.Panicf("%s: %s", "Server runtime error", err)
	}
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

	exchangeName := os.Getenv("EXCHANGE_NAME")
	if exchangeName == "" {
		panic("No value for channel name")
	}

	return ShoulderConfig{AmqpConnString: amqp, PostgresConnString: postgres, ExchangeName: exchangeName}
}

// get all records from the db and put them in memory LOL
func project(db *gorm.DB, state map[int]string) {
	var accretions []database.Accretion
	db.Find(&accretions)
	for _, value := range accretions {
		state[value.Key] = value.Value
	}
}

// listen on msgs, put messages in state, and persist
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
