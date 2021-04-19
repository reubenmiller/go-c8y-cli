package list

import (
	"net/url"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/reubenmiller/go-c8y-cli/pkg/activitylogger"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/pkg/timestamp"
	"github.com/spf13/cobra"
)

type CmdList struct {
	*subcommand.SubCommand

	entryType string
	dateFrom  string
	dateTo    string

	factory *cmdutil.Factory
}

func NewCmdList(f *cmdutil.Factory) *CmdList {
	ccmd := &CmdList{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get activity log entries",
		Long:  `View activity log entries and filter for specific information`,
		Example: heredoc.Doc(`
		$ c8y activitylog list --datFrom -1h
		Show entries from the last hour

		$ c8y activitylog list --dateFrom -8h --filter "method match PUT|POST"
		Show entries from the last 8 hours only included PUT and POST requests

		$ c8y activitylog list --filter "statusCode > 299"
		Show all failed requests
		
		$ c8y activitylog list --filter "method eq POST" --select responseTimeMS,time,ctx,responseSelf -o csv | sort -n | tail -5
		Get the top 5 slowest response times from POST requests.
		`),
		RunE: ccmd.RunE,
	}

	cmd.Flags().StringVar(&ccmd.entryType, "type", "request", "Type of entry")
	cmd.Flags().StringVar(&ccmd.dateFrom, "dateFrom", "", "Start date or date and time of the log entry")
	cmd.Flags().StringVar(&ccmd.dateTo, "dateTo", "", "End date or date and time of the log entry")

	cmd.SilenceUsage = true

	completion.WithOptions(
		cmd,
		completion.WithValidateSet("type", "command", "request", "user", "all"),
	)

	cmdutil.DisableAuthCheck(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

func (n *CmdList) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}
	activitylog, err := n.factory.ActivityLogger()

	if err != nil {
		return err
	}
	filter := activitylogger.Filter{
		Type: n.entryType,
	}
	if strings.EqualFold(n.entryType, "all") {
		// dont filter by type
		filter.Type = ""
	}

	if n.dateFrom != "" {
		dateFrom, err := timestamp.TryGetTimestamp(n.dateFrom, false)
		if err != nil {
			return err
		}
		filter.DateFrom = dateFrom
	}

	if n.dateTo != "" {
		dateTo, err := timestamp.TryGetTimestamp(n.dateTo, false)
		if err != nil {
			return err
		}
		filter.DateTo = dateTo
	}

	u, err := url.Parse(cfg.GetHost())
	if err != nil {
		filter.Host = strings.ReplaceAll(cfg.GetHost(), "https://", "")
	} else {
		filter.Host = u.Host
	}
	cfg.Logger.Debugf("activity log filter: path=%s, host=%s, datefrom=%s, dateto=%s", activitylog.GetPath(), filter.Host, filter.DateFrom, filter.DateTo)

	err = activitylog.GetLogEntries(filter, func(line []byte) error {
		return n.factory.WriteJSONToConsole(cfg, cmd, "", line)
	})

	return err
}
