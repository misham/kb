package cmd

import (
	"fmt"

	"github.com/alecthomas/kong"
)

// Version, Commit, and Date are set via -ldflags at build time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// CLI defines the top-level command structure for kb.
type CLI struct {
	DB          string           `help:"Path to database." default:"kb.db" env:"KB_DB"`
	Plain       bool             `help:"Disable styled output."`
	VersionFlag kong.VersionFlag `name:"version" short:"v" help:"Show version information."`
	Version     VersionCmd       `cmd:"" help:"Show version information."`
	Init        InitCmd          `cmd:"" help:"Create a new knowledge base."`
	Import      ImportCmd        `cmd:"" help:"Import markdown files."`
	Search      SearchCmd        `cmd:"" help:"Search the knowledge base."`
	List        ListCmd          `cmd:"" help:"List documents."`
	Get         GetCmd           `cmd:"" help:"Display a document."`
	Delete      DeleteCmd        `cmd:"" help:"Delete a document."`
	Link        LinkCmd          `cmd:"" help:"Link two documents."`
	Links       LinksCmd         `cmd:"" help:"Show linked documents."`
	MergeDriver MergeDriverCmd   `cmd:"" name:"merge-driver" help:"Merge two kb databases during git merge. Git invokes this automatically when merging *.db files. Documents are matched by (type, title); the newer version wins conflicts. Run 'kb setup-git' first to register the driver."`
	SetupGit    SetupGitCmd      `cmd:"" name:"setup-git" help:"Register kb as a git merge driver and mergetool so that 'git merge' resolves kb.db conflicts automatically. If a conflict still occurs, run 'git mergetool --tool=kb' to resolve it."`
}

// VersionCmd prints version information.
type VersionCmd struct{}

// Run prints the version string.
func (v *VersionCmd) Run(*CLI) error {
	fmt.Printf("kb %s\ncommit: %s\nbuilt:  %s\n", Version, Commit, Date)
	return nil
}
