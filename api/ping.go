package api

import (
	"encoding/json"
	"net/http"
	gen "shoulder/api/gen"
	"time"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	commandTimeHistogram = promauto.NewHistogram(prometheus.HistogramOpts{Name: "command_seconds", Buckets: prometheus.DefBuckets})
	queryTimeHistogram   = promauto.NewHistogram(prometheus.HistogramOpts{Name: "query_seconds", Buckets: prometheus.DefBuckets})
)

type ShoulderAPI struct {
	Channel      *amqp.Channel
	ExchangeName string
	State        map[int]string
}

// (POST /command)
func (api ShoulderAPI) PostCommand(c *gin.Context) {
	start := time.Now().UnixMilli()
	var command gen.CommandContent
	err := c.Bind(&command)
	if err != nil {
		log.Println("Could not bind command content: ", err)
	}
	json, err := json.Marshal(command)
	if err != nil {
		log.Println("Could not marshal command: ", err)
	}
	err = api.Channel.PublishWithContext(c, api.ExchangeName, "", false, false, amqp.Publishing{ContentType: "text/plain", Body: json})
	if err != nil {
		log.Println("Could not publish: ", err)
	}
	c.JSON(http.StatusOK, gen.CommandAccepted("Sent message"))
	commandTimeHistogram.Observe(float64(time.Now().UnixMilli()-start) / 1000)
}

// (GET /query)
func (api ShoulderAPI) GetQuery(c *gin.Context) {
	start := time.Now().UnixMilli()
	c.JSON(http.StatusOK, api.State)
	queryTimeHistogram.Observe(float64(time.Now().UnixMilli()-start) / 1000)
}
