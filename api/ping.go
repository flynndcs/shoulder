package ping

import (
	"encoding/json"
	"net/http"
	gen "shoulder/api/gen"

	"log"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

type PingAPI struct {
	Channel      *amqp.Channel
	ExchangeName string
	State        map[int]string
}

// (POST /command)
func (api PingAPI) PostCommand(c *gin.Context) {
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
}

// (GET /query)
func (api PingAPI) GetQuery(c *gin.Context) {
	c.JSON(http.StatusOK, api.State)
}
