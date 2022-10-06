package internal

import (
	"log"
	"os"
	gen "shoulder/api/gen"

	"github.com/getkin/kin-openapi/openapi3"
)

type ShoulderConfig struct {
	AmqpConnString     string
	PostgresConnString string
	ExchangeName       string
}

func GetConfig() (*openapi3.T, ShoulderConfig) {
	swagger, err := gen.GetSwagger()
	if err != nil {
		log.Panicf("%s: %s", "Failed to get swagger spec", err)
	}

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

	return swagger, ShoulderConfig{AmqpConnString: amqp, PostgresConnString: postgres, ExchangeName: exchangeName}
}
