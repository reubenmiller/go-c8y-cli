package c8ysubscribe

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

// Options subscription options
type Options struct {
	// Timeout duration
	Timeout time.Duration

	// MaxMessages maximum messages
	MaxMessages int64

	// ActionTypes filter by action types
	ActionTypes []string

	// OnMessage on message callback
	OnMessage func(msg string) error
}

// Subscribe subscribe to a single channel
func Subscribe(client *c8y.Client, log *logger.Logger, channelPattern string, opts Options) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("could not create realtime client. %s", r)
		}
	}()

	if err := client.Realtime.Connect(); err != nil {
		log.Errorf("Could not connect to /cep/realtime. %s", err)
		return err
	}

	msgCh := make(chan (*c8y.Message))

	// Enable ctrl-c stop signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	log.Infof("Listenening to subscriptions: %s", channelPattern)

	client.Realtime.Subscribe(channelPattern, msgCh)

	timeoutDuration := opts.Timeout
	timeoutCh := time.After(timeoutDuration)

	defer func() {
		// client.Realtime.UnsubscribeAll()
		_ = client.Realtime.Disconnect()
	}()

	var receivedCounter int64
	actionTypes := strings.ToLower(strings.Join(opts.ActionTypes, " "))

	for {
		select {
		case <-timeoutCh:
			log.Info("Duration has expired. Stopping realtime client")
			return nil
		case msg := <-msgCh:
			if actionTypes == "" || strings.Contains(actionTypes, strings.ToLower(msg.Payload.RealtimeAction)) {

				data := jsonUtilities.UnescapeJSON(msg.Payload.Data)
				_ = opts.OnMessage(data + "\n")

				receivedCounter++

				if opts.MaxMessages != 0 && receivedCounter >= opts.MaxMessages {
					return nil
				}
			}

		case <-signalCh:
			// Enable ctrl-c to stop
			log.Info("Stopping realtime client")
			return nil
		}
	}
}

// SubscribeMultiple subscribe to multiple channels
func SubscribeMultiple(client *c8y.Client, log *logger.Logger, channelPatterns []string, opts Options) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("could not create realtime client. %s", r)
		}
	}()

	if err := client.Realtime.Connect(); err != nil {
		log.Errorf("Could not connect to /cep/realtime. %s", err)
		return nil
	}

	msgCh := make(chan (*c8y.Message))

	// Enable ctrl-c stop signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	for _, pattern := range channelPatterns {
		log.Infof("Listenening to subscriptions: %s", pattern)

		client.Realtime.Subscribe(pattern, msgCh)
	}

	timeoutDuration := opts.Timeout
	timeoutCh := time.After(timeoutDuration)

	defer func() {
		_ = client.Realtime.Disconnect()
	}()

	var receivedCounter int64

	for {
		select {
		case <-timeoutCh:
			log.Info("Duration has expired. Stopping realtime client")
			return nil
		case msg := <-msgCh:

			data := jsonUtilities.UnescapeJSON(msg.Payload.Data)

			// return data from cli
			if opts.OnMessage != nil {
				_ = opts.OnMessage(data + "\n")
			}

			receivedCounter++

			if opts.MaxMessages != 0 && receivedCounter >= opts.MaxMessages {
				return nil
			}

		case <-signalCh:
			// Enable ctrl-c to stop
			log.Info("Stopping realtime client")
			return nil
		}
	}
}
