package main

import (
	"errors"
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
	parser, err := kong.New(&cli,
		kong.Name("kb"),
		kong.Description("A knowledge base CLI backed by SQLite."),
		kong.Help(cmd.HelpPrinter),
		kong.Vars{"version": fmt.Sprintf("kb %s\ncommit: %s\nbuilt:  %s", version, commit, date)},
	)
	if err != nil {
		panic(err)
	}

	ctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		printErrMsg(err.Error())
		var parseErr *kong.ParseError
		if errors.As(err, &parseErr) {
			_ = parseErr.Context.PrintUsage(false)
		}
		os.Exit(1)
	}
	if err := ctx.Run(&cli); err != nil {
		printErrMsg(err.Error())
		os.Exit(1)
	}
}

func printErrMsg(msg string) {
	// nosec: msg is derived from internal error strings, not raw user input.
	if cmd.PlainOutput {
		fmt.Fprintln(os.Stderr, "Error: "+msg) //nolint:gosec
	} else {
		fmt.Fprintln(os.Stderr, errStyle.Render("Error: "+msg)) //nolint:gosec
	}
}
