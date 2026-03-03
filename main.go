package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"charm.land/lipgloss/v2"

	"kb/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var errStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Red)

func main() {
	cmd.Version = version
	cmd.Commit = commit
	cmd.Date = date
	// Set PlainOutput before kong.Parse because kong's HelpPrinter fires
	// during Parse (via BeforeReset), before flag values are applied to the struct.
	for _, arg := range os.Args[1:] {
		if arg == "--plain" {
			cmd.PlainOutput = true
			break
		}
	}

	var cli cmd.CLI
	ctx := kong.Parse(&cli,
		kong.Name("kb"),
		kong.Description("A knowledge base CLI backed by SQLite."),
		kong.UsageOnError(),
		kong.Help(cmd.HelpPrinter),
		kong.Vars{"version": fmt.Sprintf("kb %s\ncommit: %s\nbuilt:  %s", version, commit, date)},
	)
	if err := ctx.Run(&cli); err != nil {
		if cmd.PlainOutput {
			fmt.Fprintln(os.Stderr, "Error: "+err.Error())
		} else {
			fmt.Fprintln(os.Stderr, errStyle.Render("Error: "+err.Error()))
		}
		os.Exit(1)
	}
}
