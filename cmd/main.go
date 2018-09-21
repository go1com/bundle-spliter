package main

import (
	"context"
	"github.com/go1com/bundle-splitter"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

func main() {
	ctx := context.Background()
	flags := bundle_splitter.NewFlags()

	// Credentials can be leaked with debug enabled.
	if *flags.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Infoln("======= Bundle-Splitter =======")
		logrus.Infof("RabbitMQ URL: %s", *flags.Url)
		logrus.Infof("RabbitMQ kind: %s", *flags.Kind)
		logrus.Infof("RabbitMQ exchange: %s", *flags.Exchange)
		logrus.Infof("RabbitMQ queue name: %s", *flags.QueueName)
		logrus.Infof("RabbitMQ consumer name: %s", *flags.ConsumerName)
		logrus.Infoln("====================================")
	}

	terminate := make(chan os.Signal, 1)

	app, err := flags.NewApplication()
	if err != nil {
		logrus.WithError(err).Panicln("app")
	}

	go app.Start(ctx, terminate)

	signal.Notify(terminate, os.Interrupt)
	<-terminate
	os.Exit(1)
}
