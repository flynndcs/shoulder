package internal

import (
	"encoding/json"
	"log"
	"net/http"
	"shoulder/api"
	database "shoulder/db"

	gen "shoulder/api/gen"

	middleware "github.com/deepmap/oapi-codegen/pkg/gin-middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

var (
	recordsInStateGauge = promauto.NewGauge(prometheus.GaugeOpts{Name: "records_count"})
	valueSizeHist       = promauto.NewHistogram(prometheus.HistogramOpts{Name: "value_size", Buckets: prometheus.LinearBuckets(0, 10, 10)})
)

func InitServer(shoulderConfig ShoulderConfig, channel *amqp.Channel, db *gorm.DB, q amqp.Queue, swagger *openapi3.T) {
	shoulderAPI := api.ShoulderAPI{Channel: channel, ExchangeName: shoulderConfig.ExchangeName, State: make(map[int]string)}
	project(db, shoulderAPI.State)

	msgs, err := shoulderAPI.Channel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Panicf("%s: %s", "Could not consume message", err)
	}

	go listen(msgs, shoulderAPI.State, db)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()

	r := gin.Default()
	r.Use(middleware.OapiRequestValidator(swagger))

	gen.RegisterHandlers(r, shoulderAPI)
	if err != nil {
		log.Panicf("%s: %s", "Could not create API", err)
	}

	err = r.Run()
	if err != nil {
		log.Panicf("%s: %s", "Server runtime error", err)
	}
}

// get all records from the db and put them in memory LOL
func project(db *gorm.DB, state map[int]string) {
	var accretions []database.Accretion
	db.Find(&accretions)
	numRecords := 0
	for _, value := range accretions {
		numRecords++
		state[value.Key] = value.Value
		valueSizeHist.Observe(float64(len(value.Value)))
	}
	recordsInStateGauge.Set(float64(numRecords))
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
			recordsInStateGauge.Inc()
			valueSizeHist.Observe(float64(len(accretion.Value)))
			db.Create(&accretion)
		} else {
			accretion.Value = accretions[0].Value + message.Value
			valueSizeHist.Observe(float64(len(accretion.Value)))
			db.Save(&accretion)
		}
	}
}
