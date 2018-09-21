package bundle_spliter

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"os"
	"time"
)

type App struct {
	con          *amqp.Connection
	read         *amqp.Channel
	write        *amqp.Channel
	queueName    string
	consumerName string
	exchange     string
	started      chan bool
}

type RelationshipObject struct {
	Type interface{} `json:"type"`
}

func (app *App) Start(ctx context.Context, terminate <-chan os.Signal) {
	messages, err := app.messages()
	if err != nil {
		logrus.WithError(err).Panicln("start")
	}

	go func() {
		for {
			select {
			case <-terminate:
				return

			case msg := <-messages:
				if msg.DeliveryTag == 0 {
					app.read.Nack(msg.DeliveryTag, false, false)
					continue
				}

				ro := &RelationshipObject{}
				err := json.Unmarshal(msg.Body, &ro)
				if err != nil {
					logrus.WithError(err).Error("[json]")
					app.read.Nack(msg.DeliveryTag, false, false)
					continue
				}

				bundle := fmt.Sprint(ro.Type)
				if bundle == "" {
					logrus.
						WithField("routingKey", bundle).
						Error("failed to detect bundle from", string(msg.Body))

					continue
				}

				err = app.write.Publish(app.exchange, msg.RoutingKey+"."+bundle, false, false, amqp.Publishing{
					Headers:         msg.Headers,
					ContentType:     msg.ContentType,
					ContentEncoding: msg.ContentEncoding,
					Body:            msg.Body,
					DeliveryMode:    msg.DeliveryMode,
					Priority:        msg.Priority,
				})

				if err != nil {
					logrus.WithError(err).Panicln("publishing")
				}

				logrus.
					WithField("bundle", bundle).
					WithField("body", string(msg.Body)).
					Debugln("dispatched to", msg.RoutingKey+"."+bundle)

				err = app.read.Ack(msg.DeliveryTag, true)
				if err != nil {
					logrus.WithError(err).Panicln("ack")
				}
			}
		}
	}()

	time.Sleep(2 * time.Second)
	app.started <- true
}

func (app *App) messages() (<-chan amqp.Delivery, error) {
	queue, err := app.read.QueueDeclare(app.queueName, false, false, false, false, nil, )

	if err != nil {
		return nil, err
	}

	routingKeys := []string{"ro.create", "ro.update", "ro.delete"}
	for _, routingKey := range routingKeys {
		err = app.read.QueueBind(queue.Name, routingKey, app.exchange, true, nil)
		if err != nil {
			return nil, err
		}
	}

	messages, err := app.read.Consume(queue.Name, app.consumerName, false, false, false, true, nil)

	return messages, err
}
