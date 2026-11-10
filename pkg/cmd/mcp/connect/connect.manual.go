package connect

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mcpserver"
	"github.com/spf13/cobra"
)

type CmdConnect struct {
	*subcommand.SubCommand

	factory *cmdutil.Factory
	socket  string
}

func NewCmdConnect(f *cmdutil.Factory) *CmdConnect {
	ccmd := &CmdConnect{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect an MCP client to a running MCP server",
		Long: heredoc.Doc(`
			Relay the stdio of an MCP client to an MCP server started with "c8y mcp serve".

			Register this command in the MCP client instead of the server. It does not use any
			credentials; it only forwards messages between the client and the server's unix socket.
			If the server is not running, the agent is told to ask the user to start it.

			The relay doesn't depend on the session. The session is selected, and can be switched,
			in the terminal running "c8y mcp serve", so the relay only needs to be registered once.
		`),
		Example: heredoc.Doc(`
			$ claude mcp add --scope user c8y -- c8y mcp connect
			Add the MCP server to Claude Code for all projects

			$ claude mcp add c8y -- c8y mcp connect --socket /path/to/server.sock
			Connect to a server which was started with a custom --socket path
		`),
		Args: cobra.NoArgs,
		RunE: ccmd.RunE,
	}

	cmd.Flags().StringVar(&ccmd.socket, "socket", "", "Unix socket path of the server. Only needed if the server was started with --socket")

	cmdutil.DisableEncryptionCheck(cmd)
	cmdutil.DisableAuthCheck(cmd)
	cmd.SilenceUsage = true
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)
	return ccmd
}

func (n *CmdConnect) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}

	socketPath := n.socket
	if socketPath == "" {
		socketPath = mcpserver.DefaultSocketPath(cfg.GetSessionHomeDir())
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	conn, message := n.dial(socketPath)
	if conn == nil {
		fmt.Fprintln(n.factory.IOStreams.ErrOut, message)
		server := mcpserver.NewUnavailableServer(n.factory.BuildVersion, message)
		if runErr := server.Run(ctx, &mcp.StdioTransport{}); runErr != nil && ctx.Err() == nil {
			return runErr
		}
		return nil
	}

	return mcpserver.Relay(ctx, conn, os.Stdin, os.Stdout)
}

// dial connects to the server socket. If it fails, it returns a message for the agent explaining what to do
func (n *CmdConnect) dial(socketPath string) (net.Conn, string) {
	// Don't connect to a socket which another user could have replaced
	if err := mcpserver.CheckSocketDir(filepath.Dir(socketPath)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Sprintf("Refusing to connect to the go-c8y-cli MCP server (socket %s) as it is not secure. Ask the user to fix it. %s", socketPath, err)
	}

	conn, err := net.DialTimeout("unix", socketPath, 5*time.Second)
	if err != nil {
		startCmd := "c8y mcp serve"
		if n.socket != "" {
			startCmd += " --socket " + n.socket
		}
		return nil, fmt.Sprintf(
			"The go-c8y-cli MCP server is not running (socket %s). Ask the user to start it in a terminal with \"%s\" and then reconnect the MCP server (e.g. /mcp in Claude Code).",
			socketPath, startCmd,
		)
	}

	// Only talk to a server run by the same user
	if err := mcpserver.CheckPeerUser(conn); err != nil {
		conn.Close()
		return nil, fmt.Sprintf("Refusing to connect to the go-c8y-cli MCP server (socket %s). %s", socketPath, err)
	}
	return conn, ""
}
