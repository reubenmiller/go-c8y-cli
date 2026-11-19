package clisurface

import (
	"strings"
	"testing"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/completion"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flags"
	"github.com/spf13/cobra"
)

func noopRunE(*cobra.Command, []string) error { return nil }

func buildTree() *cobra.Command {
	root := &cobra.Command{Use: "c8y"}
	devices := &cobra.Command{Use: "devices", Short: "Cumulocity devices"} // group, not projectable

	get := &cobra.Command{Use: "get", Short: "Get device", Long: "Get an existing device", RunE: noopRunE}
	get.Flags().StringSlice("id", []string{""}, "Device ID (required) (accepts pipeline)")
	get.Flags().Bool("withChildren", false, "include children")
	get.Flags().String("data", "", "static data") // common dynamic flag, must be excluded
	flags.WithOptions(get,
		flags.WithExtendedPipelineSupport("id", "id", true, "deviceId"),
		flags.WithPowershellName("Get-Device"),
		flags.WithOutputType("application/vnd.com.nsn.cumulocity.customDevice+json", ""),
	)

	list := &cobra.Command{Use: "list", Short: "Get device collection", RunE: noopRunE}
	list.Flags().String("availability", "", "Filter by status")
	completion.WithOptions(list, completion.WithValidateSet("availability", "AVAILABLE", "UNAVAILABLE"))
	flags.WithOptions(list, flags.WithCollectionProperty("managedObjects"))

	hidden := &cobra.Command{Use: "secret", RunE: noopRunE, Hidden: true} // must be skipped

	devices.AddCommand(get, list, hidden)
	root.AddCommand(devices)
	return root
}

func find(cmds []Command, path string) *Command {
	for i := range cmds {
		if strings.Join(cmds[i].Path, " ") == path {
			return &cmds[i]
		}
	}
	return nil
}

func TestWalk(t *testing.T) {
	cmds := Walk(buildTree(), "")

	if find(cmds, "devices") != nil {
		t.Error("group command 'devices' should not be projectable")
	}
	if find(cmds, "devices secret") != nil {
		t.Error("hidden command should be skipped")
	}

	get := find(cmds, "devices get")
	if get == nil {
		t.Fatalf("devices get not found in surface; got %d commands", len(cmds))
	}
	if get.Noun != "devices" || get.Name != "get" {
		t.Errorf("noun/name = %q/%q, want devices/get", get.Noun, get.Name)
	}
	if get.Method != "GET" {
		t.Errorf("method = %q, want GET", get.Method)
	}
	if get.PowershellName != "Get-Device" {
		t.Errorf("powershellName = %q, want Get-Device", get.PowershellName)
	}
	if get.Accept != "application/vnd.com.nsn.cumulocity.customDevice+json" {
		t.Errorf("accept = %q", get.Accept)
	}
	if get.PipelineFlag != "id" {
		t.Errorf("pipelineFlag = %q, want id", get.PipelineFlag)
	}

	// flag surface: data excluded, id is pipeline+required slice
	if f := findFlag(get, "data"); f != nil {
		t.Error("common dynamic flag 'data' must be excluded from surface")
	}
	id := findFlag(get, "id")
	if id == nil {
		t.Fatal("id flag missing")
	}
	if id.Type != "stringSlice" || !id.Pipeline || !id.Required {
		t.Errorf("id flag = %+v, want type stringSlice pipeline required", *id)
	}

	list := find(cmds, "devices list")
	if list == nil {
		t.Fatal("devices list not found")
	}
	if list.CollectionProperty != "managedObjects" {
		t.Errorf("collectionProperty = %q, want managedObjects", list.CollectionProperty)
	}
	av := findFlag(list, "availability")
	if av == nil || len(av.ValidateSet) != 2 || av.ValidateSet[0] != "AVAILABLE" {
		t.Errorf("availability validateSet = %v, want [AVAILABLE UNAVAILABLE]", av)
	}
}

func TestWalkFilter(t *testing.T) {
	cmds := Walk(buildTree(), "devices get")
	if len(cmds) != 1 || cmds[0].Name != "get" {
		t.Fatalf("filter 'devices get' = %d commands, want 1 (get)", len(cmds))
	}
	// root-prefixed filter also works
	if got := Walk(buildTree(), "c8y devices list"); len(got) != 1 || got[0].Name != "list" {
		t.Fatalf("root-prefixed filter failed: %d commands", len(got))
	}
}

func findFlag(c *Command, name string) *Flag {
	for i := range c.Flags {
		if c.Flags[i].Name == name {
			return &c.Flags[i]
		}
	}
	return nil
}
