package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	"time"

	"github.com/fatih/color"
	"github.com/reubenmiller/go-c8y-cli/pkg/jsonUtilities"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/spf13/cobra"
)

func subscribe(channelPattern string, timeoutSec int64, maxMessages int64, cmd *cobra.Command) error {

	if err := client.Realtime.Connect(); err != nil {
		Logger.Errorf("Could not connect to /cep/realtime. %s", err)
		return err
	}

	msgCh := make(chan (*c8y.Message))

	// Enable ctrl-c stop signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	Logger.Infof("Listenening to subscriptions: %s", channelPattern)

	client.Realtime.Subscribe(channelPattern, msgCh)

	timeoutDuration := time.Duration(timeoutSec) * time.Second
	timeoutCh := time.After(timeoutDuration)

	defer func() {
		// client.Realtime.UnsubscribeAll()
		client.Realtime.Disconnect()
	}()

	var receivedCounter int64

	for {
		select {
		case <-timeoutCh:
			Logger.Info("Duration has expired. Stopping realtime client")
			return nil
		case msg := <-msgCh:

			data := jsonUtilities.UnescapeJSON(msg.Payload.Data)

			// show data on console
			// cmd.Printf("%s\n", data)

			// return data from cli
			fmt.Printf("%s\n", data)

			receivedCounter++

			if maxMessages != 0 && receivedCounter >= maxMessages {
				return nil
			}

		case <-signalCh:
			// Enable ctrl-c to stop
			Logger.Info("Stopping realtime client")
			return nil
		}
	}
}

func subscribeMultiple(channelPatterns []string, timeoutSec int64, maxMessages int64, cmd *cobra.Command) error {

	if err := client.Realtime.Connect(); err != nil {
		Logger.Errorf("Could not connect to /cep/realtime. %s", err)
		return nil
	}

	msgCh := make(chan (*c8y.Message))

	// Enable ctrl-c stop signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	for _, pattern := range channelPatterns {
		Logger.Infof("Listenening to subscriptions: %s", pattern)

		client.Realtime.Subscribe(pattern, msgCh)
	}

	timeoutDuration := time.Duration(timeoutSec) * time.Second
	timeoutCh := time.After(timeoutDuration)

	defer func() {
		client.Realtime.Disconnect()
	}()

	var receivedCounter int64

	for {
		select {
		case <-timeoutCh:
			Logger.Info("Duration has expired. Stopping realtime client")
			return nil
		case msg := <-msgCh:

			data := jsonUtilities.UnescapeJSON(msg.Payload.Data)

			// set color based on channel
			channelInfo := strings.Split(msg.Channel, "/")
			// var colorPrint func(string, ...interface{}) string

			colorPrint := color.GreenString
			if len(channelInfo) > 1 {
				switch channelInfo[1] {
				case "measurements":
					colorPrint = color.GreenString
				case "events":
					colorPrint = color.HiYellowString
				case "operations":
					colorPrint = color.MagentaString
				case "alarms":
					colorPrint = color.RedString
				default:
					colorPrint = color.GreenString
				}
			}

			// show data on console
			cmd.Printf("%s%s\n", colorPrint("[%s]: ", msg.Channel), data)

			// return data from cli
			fmt.Printf("%s\n", data)

			receivedCounter++

			if maxMessages != 0 && receivedCounter >= maxMessages {
				return nil
			}

		case <-signalCh:
			// Enable ctrl-c to stop
			Logger.Info("Stopping realtime client")
			return nil
		}
	}
}
