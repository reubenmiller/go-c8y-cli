package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/internal/docs"
	"github.com/reubenmiller/go-c8y-cli/pkg/cmd"
	"github.com/spf13/pflag"
)

func main() {
	var flagError pflag.ErrorHandling

	docCmd := pflag.NewFlagSet("", flagError)
	manPage := docCmd.BoolP("man-page", "", false, "Generate manual pages")
	website := docCmd.BoolP("website", "", false, "Generate website pages")
	dir := docCmd.StringP("doc-path", "", "", "Path directory where you want generate doc files")
	help := docCmd.BoolP("help", "h", false, "Help about any command")

	if err := docCmd.Parse(os.Args); err != nil {
		os.Exit(1)
	}

	if *help {
		_, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n\n%s", os.Args[0], docCmd.FlagUsages())
		if err != nil {
			fatal(err)
		}
		os.Exit(1)
	}

	if *dir == "" {
		fatal("no dir set")
	}

	rootCmd := cmd.NewRootCmd()
	rootCmd.InitDefaultHelpCmd()
	rootCmd.ConfigureRootCmd()
	// rootCmd.ConfigureRootCmd()

	err := os.MkdirAll(*dir, 0755)
	if err != nil {
		fatal(err)
	}

	if *website {
		err = docs.GenMarkdownTreeCustom(&rootCmd.Command, *dir, filePrepender, linkHandler)
		if err != nil {
			fatal(err)
		}
	}

	if *manPage {

		header := &docs.GenManHeader{
			Title:   "c8y",
			Section: "1",
			Source:  "",
			Manual:  "",
		}
		err = docs.GenManTree(&rootCmd.Command, header, *dir)
		if err != nil {
			fatal(err)
		}
	}
}

func filePrepender(filename string, opts ...string) string {
	var category, title string
	if len(opts) >= 2 {
		category = opts[0]
		title = opts[1]
	}
	header := fmt.Sprintf(`---
layout: manual
# permalink: /:path/:basename
category: %s
title: %s
---
`, category, title)
	return header
}

func linkHandler(name string, opts ...string) string {
	return fmt.Sprintf("./%s", strings.TrimSuffix(name, ".md"))
}

func fatal(msg interface{}) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
