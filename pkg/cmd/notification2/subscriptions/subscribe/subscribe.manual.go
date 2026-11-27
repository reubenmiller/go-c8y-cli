package subscribe

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/c8ystream"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmderrors"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/config"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	apiv2 "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/notification2"
	apirealtime "github.com/reubenmiller/go-c8y/v2/pkg/c8y/api/realtime"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	realtime "github.com/reubenmiller/go-c8y/v2/pkg/c8y/notification2"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/wsdial"
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
	cmd.Flags().StringVar(&ccmd.Subscriber, "subscriber", "", "The subscriber name which the client wishes to be identified with. Defaults to token subject if not empty otherwise 'goc8ycli'")
	cmd.Flags().StringVar(&ccmd.Consumer, "consumer", "", "Consumer name. Required for shared subscriptions")
	cmd.Flags().StringSliceVar(&ccmd.ActionTypes, "actionTypes", []string{}, "Only listen for specific action types, CREATE, UPDATE or DELETE (client side filtering)")
	cmd.Flags().String("duration", "", "Subscription duration")
	cmd.Flags().Int64Var(&ccmd.ExpiresInMinutes, "expiresInMinutes", 1440, "Token expiration duration")

	completion.WithOptions(
		cmd,
		completion.WithNotification2SubscriptionName("name", ccmd.factory.ClientV2),
		completion.WithValidateSet("actionTypes", "CREATE", "UPDATE", "DELETE"),
	)

	flags.WithOptions(
		cmd,
		// The command reads a stream. The token request which it sends when no
		// token is given is an internal step, which never asks (see subscribeStream)
		flags.WithSemanticMethod("GET"),
	)

	// Required flags
	_ = cmd.MarkFlagRequired("name")

	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

// RunE subscribes to the subscription and writes the payload of every message
// (which matches the action types) until --duration passes (or the command is
// interrupted). The token is created when it is not given.
func (n *SubscribeCmd) RunE(cmd *cobra.Command, args []string) error {
	r, err := c8ystream.NewRunner(cmd, n.factory)
	if err != nil {
		return err
	}
	client, err := r.Client()
	if err != nil {
		return err
	}
	streamOpts, err := r.StreamOptions()
	if err != nil {
		return err
	}
	cfg := r.Config

	// The notification2 client logs through the default logger: show its
	// messages (the token handling, the websocket frames) only with
	// --verbose/--debug, like the logging of the SDK's client.
	level := slog.LevelWarn
	if cfg.Verbose() || cfg.Debug() {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(n.factory.IOStreams.ErrOut, &slog.HandlerOptions{Level: level}))
	// in the format of v2.59.0
	if cliLog, logErr := n.factory.Logger(); logErr == nil && cliLog != nil {
		logger = slog.New(cmdutil.NewWebsocketLogHandler(cliLog, nil, "unsubscribing", "Worker is shutting down"))
	}

	return r.Run(func(in *c8ystream.Resolver) (c8ystream.Call, error) {
		clientOpts := notification2.ClientOptions{
			Consumer: n.Consumer,
			Token:    n.Token,
			Options: notification2.TokenOptions{
				Subscriber:        n.Subscriber,
				Subscription:      n.Subscription,
				ExpiresInMinutes:  n.ExpiresInMinutes,
				DefaultSubscriber: "goc8ycli",
			},
			ConnectionOptions: realtime.ConnectionOptions{
				PingInterval: 60 * time.Second,
				PongWait:     180 * time.Second,
				WriteWait:    60 * time.Second,
				Insecure:     cfg.SkipSSLVerify(),
			},
		}
		return func(ctx context.Context) any {
			return n.messages(ctx, cfg, logger, client, clientOpts, streamOpts)
		}, nil
	})
}

// messages returns the payloads of the messages of the subscription. Every
// message is acknowledged after it was handled. The subscription is closed
// when the duration passed, the context was cancelled, or the consumer stopped
// reading.
func (n *SubscribeCmd) messages(ctx context.Context, cfg *config.Config, logger *slog.Logger, client *apiv2.Client, clientOpts notification2.ClientOptions, streamOpts apirealtime.StreamOptions) output.Seq {
	return func(yield func(jsondoc.JSONDoc, error) bool) {
		previous := slog.Default()
		slog.SetDefault(logger)
		defer slog.SetDefault(previous)

		if streamOpts.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, streamOpts.Timeout)
			defer cancel()
		}

		// Ctrl-C stops the subscription (v2.59.0 exited without an error), also
		// while it connects
		ctx, stopInterrupt := signal.NotifyContext(ctx, os.Interrupt)
		defer stopInterrupt()

		stream, err := subscribeStream(ctx, client, clientOpts, cmdutil.NewNotification2Connector(client, cfg.SkipSSLVerify()))
		if err != nil {
			if errors.Is(err, context.Canceled) {
				err = cmderrors.NewUserErrorWithExitCode(cmderrors.ExitCancel, "Interrupted")
			}
			yield(jsondoc.Empty(), err)
			return
		}
		defer stream.Close()

		for msg, err := range stream.Items() {
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				yield(jsondoc.Empty(), err)
				return
			}
			cfg.Logger.Infof("Received message: (id=%s, action=%s, description=%s) %s", msg.Identifier, msg.Action, msg.Description, msg.Data.Raw())

			if n.matchesAction(msg.Action) {
				if payload := msg.Data.Raw(); json.Valid(payload) {
					if !yield(msg.Data, nil) {
						return
					}
				} else {
					cfg.Logger.Warnf("Could not process line. only json lines are accepted. %s", payload)
				}
			}

			if err := stream.AcknowledgeMessage(msg.Identifier); err != nil {
				cfg.Logger.Warnf("Failed to send ack. %s", err)
			}
		}
	}
}

// subscribeStream creates the client of the subscription (which creates the
// token, unless one is given) and connects it with the connector. It is
// client.Notification2.SubscribeStream, except that the connector decides how
// to connect: like v2.59.0, a dry run (--dry) shows the requests which create
// the token and then connects, and the insecure setting applies to the
// certificate of the server.
func subscribeStream(ctx context.Context, client *apiv2.Client, opts notification2.ClientOptions, connector wsdial.Connector) (*notification2.MessageStream, error) {
	// The token request (and its renewal) is an internal step of the
	// subscription, not an action of the user: as in v2.59.0 it never asks for
	// a confirmation, also not with --confirm. Only the approver is removed, a
	// dry run still shows the request.
	nc, err := client.Notification2.CreateClient(apiv2.WithRequestApprover(ctx, nil), opts)
	if err != nil {
		if transportErr, ok := c8ystream.TransportError(err); ok {
			// The token request got no response (e.g. the certificate of the
			// server could not be verified). As in v2.59.0 this is a command
			// error with the error of the http client, not a server error.
			return nil, cmderrors.CommandError{
				ErrorType: cmderrors.ErrTypeCommand,
				Message:   transportErr.Message,
				ExitCode:  cmderrors.ExitError,
			}
		}
		// The token request failed with an HTTP error: as in v2.59.0 a command
		// error "METHOD url: status type message" (exit code 100), whatever the
		// status.
		if httpErr, ok := c8ystream.HTTPCommandError(err); ok {
			return nil, httpErr
		}
		return nil, err
	}
	nc.SetConnector(connector)
	if err := nc.ConnectContext(ctx); err != nil {
		_ = nc.Close()
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		// The websocket could not be opened (e.g. the handshake was rejected):
		// as in v2.59.0 a command error with exit code 100.
		return nil, cmderrors.CommandError{
			ErrorType: cmderrors.ErrTypeCommand,
			Message:   err.Error(),
			ExitCode:  cmderrors.ExitError,
		}
	}

	// The client delivers the messages to ch. Forward them to the stream until
	// the stream is closed or ctx is done.
	streamCtx, stopForwarding := context.WithCancel(ctx)
	messages := make(chan *realtime.Message, 10)
	ch := make(chan realtime.Message)
	nc.Register("*", ch)
	go func() {
		defer close(messages)
		for {
			select {
			case <-streamCtx.Done():
				return
			case msg := <-ch:
				select {
				case messages <- &msg:
				case <-streamCtx.Done():
					return
				}
			}
		}
	}()

	return realtime.NewStream(streamCtx, nc, messages, func(msg *realtime.Message) realtime.StreamData[jsondoc.JSONDoc] {
		return realtime.StreamData[jsondoc.JSONDoc]{
			Action:      msg.Action,
			Identifier:  msg.Identifier,
			Description: msg.Description,
			Data:        jsondoc.New(msg.Payload),
		}
	}, func() {
		stopForwarding()
		if err := nc.Unsubscribe(); err != nil {
			slog.Warn("Failed to unsubscribe", "err", err)
		}
		_ = nc.Close()
	}), nil
}

// matchesAction reports whether the action matches the action types (every
// action matches when no action type is given)
func (n *SubscribeCmd) matchesAction(action string) bool {
	if len(n.ActionTypes) == 0 {
		return true
	}
	return slices.ContainsFunc(n.ActionTypes, func(item string) bool {
		return strings.EqualFold(item, action)
	})
}
