package bundle_spliter

import (
	"flag"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"os"
)

type Flags struct {
	Url          *string
	Kind         *string
	Exchange     *string
	RoutingKeys  *string
	QueueName    *string
	ConsumerName *string
	Debug        *bool
}

func env(key string, defaultValue string) string {
	value, _ := os.LookupEnv(key)

	if "" == value {
		return defaultValue
	}

	return value
}

func NewFlags() Flags {
	f := Flags{}
	f.Url = flag.String("url", env("RABBITMQ_URL", "amqp://go1:go1@127.0.0.1:5672/"), "")
	f.Kind = flag.String("kind", env("RABBITMQ_KIND", "topic"), "")
	f.Exchange = flag.String("exchange", env("RABBITMQ_EXCHANGE", "events"), "")
	f.RoutingKeys = flag.String("routing-keys", env("RABBITMQ_ROUTING_KEYS", "ro.create,ro.update,ro.delete"), "")
	f.QueueName = flag.String("queue-name", env("RABBITMQ_QUEUE_NAME", "bundle-spliter"), "")
	f.ConsumerName = flag.String("consumer-name", env("RABBITMQ_CONSUMER_NAME", "bundle-spliter"), "")
	f.Debug = flag.Bool("debug", false, "Enable with care; credentials can be leaked if this is on.")
	flag.Parse()

	return f
}

func (f *Flags) QueueConnection() (*amqp.Connection, error) {
	url := *f.Url
	con, err := amqp.Dial(url)
	if nil != err {
		return nil, err
	}

	go func() {
		conCloseChan := con.NotifyClose(make(chan *amqp.Error))

		select
		{
		case err := <-conCloseChan:
			if err != nil {
				logrus.WithError(err).Panicln("connection")
			}
		}
	}()

	return con, nil
}

func (f *Flags) QueueChannel(con *amqp.Connection, consumerName string) (*amqp.Channel, error) {
	ch, err := con.Channel()
	if nil != err {
		return nil, err
	}

	err = ch.ExchangeDeclare(*f.Exchange, *f.Kind, false, false, false, false, nil)
	if nil != err {
		ch.Close()

		return nil, err
	}

	err = ch.Qos(1, 0, false)
	if nil != err {
		ch.Close()

		return nil, err
	}

	return ch, nil
}

func (f *Flags) NewApplication() (*App, error) {
	con, err := f.QueueConnection()
	if err != nil {
		return nil, err
	}

	read, err := f.QueueChannel(con, "")
	if err != nil {
		return nil, err
	}

	write, err := f.QueueChannel(con, "")
	if err != nil {
		return nil, err
	}

	app := &App{
		con:          con,
		read:         read,
		write:        write,
		queueName:    *f.QueueName,
		consumerName: *f.ConsumerName,
		exchange:     *f.Exchange,
		started:      make(chan bool, 1),
	}

	return app, nil
}
