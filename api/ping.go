package ping

import (
	"net/http"
	ping "shoulder/api/gen"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/labstack/gommon/log"
	amqp "github.com/rabbitmq/amqp091-go"
)

type PingAPI struct {
	Channel *amqp.Channel
	Queue   amqp.Queue
}

// (POST /command)
func (api PingAPI) PostCommand(c *gin.Context) {
	var command ping.CommandContent
	err := c.Bind(&command)
	if err != nil {
		log.Info("Could not bind command content")
	}
	err = api.Channel.PublishWithContext(c, "", api.Queue.Name, false, false, amqp.Publishing{ContentType: "text/plain", Body: []byte(strconv.Itoa(int(command.Key)) + ": " + command.Value)})
	if err != nil {
		log.Info("Could not publish")
	}
	c.JSON(http.StatusOK, ping.CommandAccepted("Sent message"))
}

// (GET /query)
func (api PingAPI) GetQuery(c *gin.Context) {
	msgs, err := api.Channel.Consume(api.Queue.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Info("Could not consumer")
	}
	msg := <-msgs
	c.JSON(http.StatusOK, ping.State(msg.Body))
}
