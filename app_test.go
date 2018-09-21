package bundle_splitter

import (
	"context"
	"go1/vendor/github.com/streadway/amqp"
	"os"
	"syscall"
	"testing"
)

func NewTestFlags() Flags {
	f := Flags{}
	Url := "amqp://go1:go1@127.0.0.1:5672/"
	f.Url = &Url
	Kind := "topic"
	f.Kind = &Kind
	Exchange := "events"
	f.Exchange = &Exchange
	QueueName := "bundle-splitter"
	f.QueueName = &QueueName
	ConsumerName := "bundle-splitter"
	f.ConsumerName = &ConsumerName
	Debug := true
	f.Debug = &Debug

	return f
}

func TestStringBundleName(t *testing.T) {
	ctx := context.Background()
	flags := NewTestFlags()
	terminate := make(chan os.Signal, 1)
	defer func() { terminate <- syscall.SIGKILL }()
	app, _ := flags.NewApplication()

	newMessage := make(chan amqp.Delivery)
	go func() {
		ch, _ := flags.QueueChannel(app.con, "")
		suffix := "_" + t.Name()
		queueName := app.queueName + suffix
		queue, _ := ch.QueueDeclare(queueName, false, false, false, false, nil)
		ch.QueueBind(queue.Name, "ro.create.500", app.exchange, true, nil)
		messages, _ := ch.Consume(queue.Name, queue.Name, true, false, false, true, nil)

		for m := range messages {
			newMessage <- m
		}
	}()

	go app.Start(ctx, terminate)
	<-app.started
	app.read.QueuePurge(app.queueName, false)
	app.write.QueuePurge(app.queueName, false)

	body := []byte(`{"type": "500", "source_id": 111, "target_id": "222"}`)
	app.write.Publish(app.exchange, "ro.create", false, false, amqp.Publishing{Body: body})
	msg := <-newMessage

	if "ro.create.500" != msg.RoutingKey {
		t.Error("invalid routing key")
	} else if len(body) != len(msg.Body) {
		t.Error("invalid message body")
	}
}

func TestNumericBundleName(t *testing.T) {
	ctx := context.Background()
	flags := NewTestFlags()
	terminate := make(chan os.Signal, 1)
	defer func() { terminate <- syscall.SIGKILL }()
	app, _ := flags.NewApplication()

	newMessage := make(chan amqp.Delivery)
	go func() {
		ch, _ := flags.QueueChannel(app.con, "")
		suffix := "_" + t.Name()
		queueName := app.queueName + suffix
		ch.QueuePurge(queueName, false)
		queue, _ := ch.QueueDeclare(queueName, false, false, false, false, nil)
		ch.QueueBind(queue.Name, "ro.create.111", app.exchange, true, nil)
		messages, _ := ch.Consume(queue.Name, queue.Name, true, false, false, true, nil)

		for m := range messages {
			newMessage <- m
		}
	}()

	go app.Start(ctx, terminate)
	<-app.started
	app.read.QueuePurge(app.queueName, false)
	app.write.QueuePurge(app.queueName, false)

	// With type is an int
	body := []byte(`{"type": 111, "source_id": 222, "target_id": "333"}`)
	app.write.Publish(app.exchange, "ro.create", false, false, amqp.Publishing{Body: body})
	msg := <-newMessage
	if "ro.create.111" != msg.RoutingKey {
		t.Error("invalid routing key")
	} else if len(body) != len(msg.Body) {
		t.Error("invalid message body")
	}
}
