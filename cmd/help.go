package cmd

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kong"

	"charm.land/lipgloss/v2"
)

// PlainOutput disables styled output when true. Set before kong.Parse.
var PlainOutput bool

var (
	helpHeaderStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	helpCmdStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#04B575"))
	helpFlagStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6F61"))
	helpDescStyle    = lipgloss.NewStyle().Faint(true)
	helpSectionStyle = lipgloss.NewStyle().Bold(true).Underline(true).MarginTop(1)
)

var commandExamples = map[string][]string{
	"init":   {"kb init", "kb --db ~/notes.db init"},
	"import": {"kb import notes.md -t research", "kb import *.md -t plan"},
	"search": {"kb search goroutines", "kb search sqlite -t research", `kb search '"exact phrase"'`},
	"list":   {"kb list", "kb list -t research"},
	"get":    {"kb get 1"},
	"delete": {"kb delete 1", "kb delete 1 -f"},
	"link":   {"kb link 1 2", "kb link 1 2 -r related"},
	"links":  {"kb links 1"},
}

// render applies a lipgloss style or returns plain text based on PlainOutput.
func render(s lipgloss.Style, text string) string {
	if PlainOutput {
		return text
	}
	return s.Render(text)
}

// HelpPrinter renders help output, styled or plain depending on PlainOutput.
func HelpPrinter(_ kong.HelpOptions, ctx *kong.Context) error {
	selected := ctx.Selected()
	if selected == nil {
		return printAppHelp(ctx)
	}
	return printCommandHelp(ctx, selected)
}

func printAppHelp(ctx *kong.Context) error {
	app := ctx.Model

	fmt.Fprintln(ctx.Stdout, render(helpHeaderStyle, app.Name)+" — "+app.Help)
	fmt.Fprintln(ctx.Stdout)
	fmt.Fprintln(ctx.Stdout, render(helpSectionStyle, "Usage"))
	fmt.Fprintf(ctx.Stdout, "  %s <command> [flags]\n", app.Name)

	fmt.Fprintln(ctx.Stdout)
	fmt.Fprintln(ctx.Stdout, render(helpSectionStyle, "Commands"))

	for _, cmd := range app.Leaves(true) {
		name := render(helpCmdStyle, fmt.Sprintf("  %-10s", cmd.Name))
		fmt.Fprintf(ctx.Stdout, "%s %s\n", name, cmd.Help)
	}

	flags := app.AllFlags(true)
	if len(flags) > 0 {
		fmt.Fprintln(ctx.Stdout)
		fmt.Fprintln(ctx.Stdout, render(helpSectionStyle, "Flags"))
		printFlags(ctx, flags)
	}

	fmt.Fprintln(ctx.Stdout)
	fmt.Fprintln(ctx.Stdout, render(helpSectionStyle, "Examples"))
	fmt.Fprintln(ctx.Stdout, "  kb init")
	fmt.Fprintln(ctx.Stdout, "  kb import notes.md plan.md -t research")
	fmt.Fprintln(ctx.Stdout, "  kb search goroutines -t research")
	fmt.Fprintln(ctx.Stdout, "  kb list -t plan")
	fmt.Fprintln(ctx.Stdout, "  kb link 1 2 -r related")

	fmt.Fprintln(ctx.Stdout)
	hint := fmt.Sprintf("Run \"%s <command> --help\" for more information on a command.", app.Name)
	fmt.Fprintln(ctx.Stdout, render(helpDescStyle, hint))
	return nil
}

func printCommandHelp(ctx *kong.Context, cmd *kong.Command) error {
	app := ctx.Model

	fmt.Fprintln(ctx.Stdout, render(helpHeaderStyle, cmd.FullPath())+" — "+cmd.Help)
	fmt.Fprintln(ctx.Stdout)
	fmt.Fprintln(ctx.Stdout, render(helpSectionStyle, "Usage"))
	fmt.Fprintf(ctx.Stdout, "  %s %s\n", app.Name, cmd.Summary())

	if len(cmd.Positional) > 0 {
		fmt.Fprintln(ctx.Stdout)
		fmt.Fprintln(ctx.Stdout, render(helpSectionStyle, "Arguments"))
		for _, pos := range cmd.Positional {
			name := render(helpCmdStyle, fmt.Sprintf("  %-10s", "<"+pos.Name+">"))
			fmt.Fprintf(ctx.Stdout, "%s %s\n", name, pos.Help)
		}
	}

	flags := cmd.AllFlags(true)
	if len(flags) > 0 {
		fmt.Fprintln(ctx.Stdout)
		fmt.Fprintln(ctx.Stdout, render(helpSectionStyle, "Flags"))
		printFlags(ctx, flags)
	}

	if examples, ok := commandExamples[cmd.Name]; ok {
		fmt.Fprintln(ctx.Stdout)
		fmt.Fprintln(ctx.Stdout, render(helpSectionStyle, "Examples"))
		for _, ex := range examples {
			fmt.Fprintf(ctx.Stdout, "  %s\n", ex)
		}
	}

	return nil
}

func printFlags(ctx *kong.Context, flags [][]*kong.Flag) {
	for _, group := range flags {
		for _, flag := range group {
			if flag.Hidden {
				continue
			}
			var parts []string
			if flag.Short != 0 {
				parts = append(parts, render(helpFlagStyle, "-"+string(flag.Short)+","))
			}
			parts = append(parts, render(helpFlagStyle, "--"+flag.Name))
			name := strings.Join(parts, " ")

			help := flag.Help
			if flag.HasDefault {
				help += render(helpDescStyle, fmt.Sprintf(" (default: %v)", flag.Default))
			}
			fmt.Fprintf(ctx.Stdout, "  %-30s %s\n", name, help)
		}
	}
}
