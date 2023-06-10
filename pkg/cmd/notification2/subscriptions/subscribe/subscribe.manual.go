package subscribe

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/go-c8y/pkg/c8y/notification2"
	"github.com/spf13/cobra"
)

// SubscribeCmd command
type SubscribeCmd struct {
	*subcommand.SubCommand

	Token            string
	Subscription     string
	Consumer         string
	Subscriber       string
	ActionTypes      []string
	ExpiresInMinutes int64
	Duration         time.Time

	factory *cmdutil.Factory
}

// NewSubscribeCmd creates a command to subscribe to notifications from a subscription
func NewSubscribeCmd(f *cmdutil.Factory) *SubscribeCmd {
	ccmd := &SubscribeCmd{
		factory: f,
	}
	cmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe to a subscription",
		Long: heredoc.Doc(`
			Subscribe to an existing subscription. If no token is provided, a token will be created automatically before starting the realtime client
		`),
		Example: heredoc.Doc(`
			$ c8y notification2 subscriptions subscribe --name registration
			Start listening to a subscription name registration

			$ c8y notification2 subscriptions subscribe --name registration --actionTypes CREATE --actionTypes UPDATE
			Start listening to a subscription name registration but only include CREATE and UPDATE action types (ignoring DELETE)

			$ c8y notification2 subscriptions subscribe --name registration --duration 10min
			Subscribe to a subscription for 10mins then exit

			$ c8y notification2 subscriptions create --name registration --context tenant --apiFilter managedobjects
			$ c8y notification2 subscriptions subscribe --name registration --consumer client01
			Create a subscription to all managed objects and subscribe to notifications

			$ c8y notification2 subscriptions subscribe --name registration --token "ey123123123123"
			Subscribe using a given token (instead of generating a token)
        `),
		RunE: ccmd.RunE,
	}

	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&ccmd.Subscription, "name", "", "The subscription name. Each subscription is identified by a unique name within a specific context")
	cmd.Flags().StringVar(&ccmd.Token, "token", "", "Token for the subscription. If not provided, then a token will be created")
	cmd.Flags().StringVar(&ccmd.Subscriber, "subscriber", "goc8ycli", "The subscriber name which the client wishes to be identified with. Defaults to goc8ycli")
	cmd.Flags().StringVar(&ccmd.Consumer, "consumer", "", "Consumer name. Required for shared subscriptions")
	cmd.Flags().StringSliceVar(&ccmd.ActionTypes, "actionTypes", []string{}, "Only listen for specific action types, CREATE, UPDATE or DELETE (client side filtering)")
	cmd.Flags().String("duration", "", "Subscription duration")
	cmd.Flags().Int64Var(&ccmd.ExpiresInMinutes, "expiresInMinutes", 1440, "Token expiration duration")

	completion.WithOptions(
		cmd,
		completion.WithNotification2SubscriptionName("name", func() (*c8y.Client, error) { return ccmd.factory.Client() }),
		completion.WithValidateSet("actionTypes", "CREATE", "UPDATE", "DELETE"),
	)

	flags.WithOptions(
		cmd,
	)

	// Required flags
	_ = cmd.MarkFlagRequired("name")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE executes the command
func (n *SubscribeCmd) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	client, err := n.factory.Client()
	if err != nil {
		return err
	}

	duration, err := flags.GetDurationFlag(cmd, "duration", true, time.Second)
	if err != nil {
		return err
	}

	notification2.SetLogger(cfg.Logger)
	realtime, err := client.Notification2.CreateClient(context.Background(), c8y.Notification2ClientOptions{
		Consumer: n.Consumer,
		Token:    n.Token,
		Options: c8y.Notification2TokenOptions{
			Subscriber:       n.Subscriber,
			Subscription:     n.Subscription,
			ExpiresInMinutes: n.ExpiresInMinutes,
		},
		ConnectionOptions: notification2.ConnectionOptions{
			PingInterval: 60 * time.Second,
			PongWait:     180 * time.Second,
			WriteWait:    60 * time.Second,
			Insecure:     cfg.SkipSSLVerify(),
		},
	})

	if err != nil {
		return err
	}

	if err := realtime.Connect(); err != nil {
		return err
	}

	messagesCh := make(chan notification2.Message)
	realtime.Register("*", messagesCh)

	// Enable ctrl-c stop signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	if duration > 0 {
		// Stop after a duration
		go func() {
			<-time.After(duration)
			signalCh <- os.Interrupt
		}()
	}

	var isMatch bool
	var msgAction string
	for {
		select {
		case msg := <-messagesCh:
			cfg.Logger.Infof("Received message: (action=%s, description=%s) %s", msg.Action, msg.Description, msg.Payload)

			if len(n.ActionTypes) == 0 {
				isMatch = true
			} else {
				isMatch = false
				msgAction = string(msg.Action)
				for _, item := range n.ActionTypes {
					if item == msgAction {
						isMatch = true
						break
					}
				}
			}

			if isMatch {
				if err := n.factory.WriteJSONToConsole(cfg, cmd, "", msg.Payload); err != nil {
					cfg.Logger.Warnf("Could not process line. only json lines are accepted. %s", err)
				}
			}

			if err := realtime.SendMessageAck(msg.Identifier); err != nil {
				cfg.Logger.Warnf("Failed to send ack. %s", err)
			}
		case <-signalCh:
			realtime.Close()
			return nil
		}
	}
}
