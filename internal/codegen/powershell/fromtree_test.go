package powershell

import (
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/spf13/cobra"
)

func noopRunE(*cobra.Command, []string) error { return nil }

// buildDevicesTree builds a minimal root with two migrated-style device
// commands carrying the inversion annotations (powershell name, output type,
// validate set, pipeline support).
func buildDevicesTree() *cobra.Command {
	root := &cobra.Command{Use: "c8y"}
	devices := &cobra.Command{Use: "devices", Short: "Cumulocity devices"}

	get := &cobra.Command{
		Use:     "get",
		Short:   "Get device",
		Long:    "Get an existing device",
		Example: "$ c8y devices get --id 12345\nGet device by id",
		RunE:    noopRunE,
	}
	get.Flags().StringSlice("id", []string{""}, "Device ID (required) (accepts pipeline)")
	get.Flags().Bool("withChildren", false, "Determines if children with ID and name should be returned")
	flags.WithOptions(get,
		flags.WithExtendedPipelineSupport("id", "id", true, "deviceId", "source.id"),
		flags.WithPowershellName("Get-Device"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.customDevice+json", ""),
	)

	list := &cobra.Command{
		Use:   "list",
		Short: "Get device collection",
		RunE:  noopRunE,
	}
	list.Flags().String("query", "", "Additional query filter (accepts pipeline)")
	list.Flags().String("availability", "", "Filter by c8y_Availability.status")
	completion.WithOptions(list,
		completion.WithValidateSet("availability", "AVAILABLE", "UNAVAILABLE", "MAINTENANCE"),
	)
	flags.WithOptions(list,
		flags.WithExtendedPipelineSupport("query", "query", false, "c8y_DeviceQueryString"),
		flags.WithCollectionProperty("managedObjects"),
		flags.WithPowershellName("Get-DeviceCollection"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.managedobjectcollection+json", "application/vnd.com.nsn.cumulocity.customDevice+json"),
	)

	devices.AddCommand(get, list)
	root.AddCommand(devices)
	return root
}

func TestGenerateFromTree(t *testing.T) {
	files, err := GenerateFromTree(buildDevicesTree(), "devices")
	if err != nil {
		t.Fatalf("GenerateFromTree: %v", err)
	}

	got := map[string]string{}
	for _, f := range files {
		got[f.RelPath] = string(f.Content)
	}

	gd, ok := got["Public/Get-Device.ps1"]
	if !ok {
		t.Fatalf("Get-Device.ps1 not generated; got %v", keys(got))
	}
	for _, want := range []string{
		"Function Get-Device {",
		"Get an existing device",
		`Type = "application/vnd.com.nsn.cumulocity.customDevice+json"`,
		`Get-ClientCommonParameters -Type "Get"`,
		"$Id",
		"[object[]]", // pipeline id flag becomes object[]
		"Mandatory = $true",
		"ValueFromPipeline=$true",
		"c8y devices get $c8yargs",
	} {
		if !strings.Contains(gd, want) {
			t.Errorf("Get-Device.ps1 missing %q", want)
		}
	}
	if strings.Contains(gd, "$WithChildren ") && !strings.Contains(gd, "[switch]") {
		t.Errorf("bool flag should render as [switch]")
	}

	gc, ok := got["Public/Get-DeviceCollection.ps1"]
	if !ok {
		t.Fatalf("Get-DeviceCollection.ps1 not generated; got %v", keys(got))
	}
	for _, want := range []string{
		"Function Get-DeviceCollection {",
		"[ValidateSet('AVAILABLE','UNAVAILABLE','MAINTENANCE')]", // validate-set annotation round-trips
		`Get-ClientCommonParameters -Type "Get", "Collection"`,   // collection accept -> Collection set
		`ItemType = "application/vnd.com.nsn.cumulocity.customDevice+json"`,
		"c8y devices list $c8yargs",
	} {
		if !strings.Contains(gc, want) {
			t.Errorf("Get-DeviceCollection.ps1 missing %q", want)
		}
	}
}

func TestDerivePowershellName(t *testing.T) {
	cases := []struct {
		noun, verb, want string
	}{
		{"devices", "get", "Get-Device"},
		{"devices", "list", "Get-DeviceCollection"},
		{"alarms", "create", "New-Alarm"},
		{"events", "update", "Update-Event"},
		{"operations", "delete", "Remove-Operation"},
		{"device groups", "get", "Get-DeviceGroup"}, // multi-word noun uses the tail
	}
	for _, c := range cases {
		if got := derivePowershellName(c.noun, c.verb); got != c.want {
			t.Errorf("derivePowershellName(%q,%q) = %q, want %q", c.noun, c.verb, got, c.want)
		}
	}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
