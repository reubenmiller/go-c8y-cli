package discover

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/grandcat/zeroconf"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmd/subcommand"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/cmdutil"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/mdns"
	"github.com/spf13/cobra"
)

type CmdDiscover struct {
	*subcommand.SubCommand

	Type   []string
	Domain string

	factory *cmdutil.Factory
}

func NewCmdDiscover(f *cmdutil.Factory) *CmdDiscover {
	ccmd := &CmdDiscover{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan for local devices using mdns-sd",
		Long:  `Generic utility to scan for devices in a given domain which respond to a mdns service discovery multicast request`,
		Example: heredoc.Doc(`
			$ echo myfile.json | c8y util scan --select id,name
			Process input json lines files and select id and name fields

			$ c8y devices list > devices.json
			$ c8y util show --input devices.json --select id,name --output csv
			Save a devices list to file, then process the file in a second step and convert it to csv only keeping id and name columns (with no headers)  
		`),
		RunE: ccmd.RunE,
	}

	cmd.Flags().StringSliceVar(&ccmd.Type, "type", []string{"_thin-edge_mqtt._tcp"}, "input value to be repeated (required) (accepts pipeline)")
	cmd.Flags().StringVar(&ccmd.Domain, "domain", "local", "Domain")

	cmdutil.DisableEncryptionCheck(cmd)
	cmd.SilenceUsage = true

	flags.WithOptions(
		cmd,
		flags.WithExtendedPipelineSupport("input", "input", true),
	)

	cmdutil.DisableAuthCheck(cmd)
	ccmd.SubCommand = subcommand.NewSubCommand(cmd)

	return ccmd
}

type Service struct {
	Host            string   `json:"hostname,omitempty"`
	Instance        string   `json:"instance,omitempty"`
	IPaddr4         string   `json:"ipv4,omitempty"`
	IPaddr6         string   `json:"ipv6,omitempty"`
	Text            []string `json:"txt,omitempty"`
	Domain          string   `json:"domain,omitempty"`
	Service         string   `json:"service,omitempty"`
	ServiceTypeName string   `json:"serviceTypeName,omitempty"`
}

func (n *CmdDiscover) RunE(cmd *cobra.Command, args []string) error {
	cfg, err := n.factory.Config()
	if err != nil {
		return err
	}

	duration := cfg.RequestTimeout()
	if duration == 0 {
		duration = 5 * time.Second
	}

	var wg sync.WaitGroup

	for _, serviceType := range n.Type {
		wg.Add(1)
		entries := make(chan *zeroconf.ServiceEntry)
		go func(results <-chan *zeroconf.ServiceEntry) {
			for entry := range results {
				hostname := strings.TrimRight(entry.HostName, ".")

				data := Service{
					Instance:        strings.ReplaceAll(entry.Instance, "\\", ""),
					Host:            hostname,
					Text:            entry.Text,
					Domain:          entry.Domain,
					Service:         entry.Service,
					ServiceTypeName: entry.ServiceTypeName(),
				}
				if len(entry.AddrIPv4) > 0 {
					data.IPaddr4 = entry.AddrIPv4[0].String()
				}

				if len(entry.AddrIPv6) > 0 {
					data.IPaddr6 = entry.AddrIPv6[0].String()
				}

				b, err := json.Marshal(data)
				if err == nil {
					if err := n.factory.WriteOutputWithoutPropertyGuess(b, cmdutil.OutputContext{}); err != nil {
						cfg.Logger.Warnf("Could not process line. only json lines are accepted. %s", err)
					}
				} else {
					cfg.Logger.Warnf("Could not marshal response to json. %s", err)
				}
			}
		}(entries)

		go func() {
			cfg.Logger.Debugf("Browsing for mdns-sd entries. type=%s, domain=%s", serviceType, n.Domain)
			mdns.Discover(serviceType, n.Domain, duration, entries)
			wg.Done()
		}()
	}

	wg.Wait()
	return nil

	// inputIterators, err := cmdutil.NewRequestInputIterators(cmd, cfg)
	// if err != nil {
	// 	return err
	// }

	// var iter iterator.Iterator
	// _, input, err := flags.WithPipelineIterator(&flags.PipelineOptions{
	// 	Name:        "input",
	// 	InputFilter: flags.FilterJsonLines,
	// 	Disabled:    inputIterators.PipeOptions.Disabled,
	// 	Required:    true,
	// })(cmd, inputIterators)

	// if err != nil {
	// 	return &flags.ParameterError{
	// 		Name: "input",
	// 		Err:  fmt.Errorf("Missing required parameter or pipeline input. %w", flags.ErrParameterMissing),
	// 	}
	// }

	// switch v := input.(type) {
	// case iterator.Iterator:
	// 	iter = v
	// default:
	// 	// use a single input iterator
	// 	iter = iterator.NewRepeatIterator("", 1)
	// }

	// bounded := iter.IsBound()
	// for {
	// 	responseText, _, err := iter.GetNext()
	// 	if err != nil {
	// 		if err == io.EOF {
	// 			break
	// 		}
	// 		return err
	// 	}

	// 	if !jsonUtilities.IsJSONObject(responseText) {
	// 		cfg.Logger.Warnf("Could not process line. only json lines are accepted")
	// 		continue
	// 	}

	// 	if err := n.factory.WriteOutputWithoutPropertyGuess(responseText, cmdutil.OutputContext{}); err != nil {
	// 		cfg.Logger.Warnf("Could not process line. only json lines are accepted. %s", err)
	// 	}

	// 	if !bounded {
	// 		break
	// 	}
	// }

	// return nil
}
