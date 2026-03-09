package cmd

import (
	"fmt"
	"os"

	"kb/internal/store"
)

// MergeDriverCmd is a git merge driver for kb database files.
// Git invokes this as: kb merge-driver %O %A %B
// where %O=ancestor, %A=ours (also the output), %B=theirs.
type MergeDriverCmd struct {
	Ancestor string `arg:"" help:"Path to ancestor version."`
	Ours     string `arg:"" help:"Path to ours version (also the merge output)."`
	Theirs   string `arg:"" help:"Path to theirs version."`
}

// Run executes the merge driver.
func (c *MergeDriverCmd) Run(_ *CLI) error {
	oursStore, err := store.Open(c.Ours)
	if err != nil {
		return fmt.Errorf("opening ours db: %w", err)
	}
	defer oursStore.Close()

	theirsStore, err := store.Open(c.Theirs)
	if err != nil {
		return fmt.Errorf("opening theirs db: %w", err)
	}
	defer theirsStore.Close()

	if err := store.MergeDB(oursStore, theirsStore); err != nil {
		return fmt.Errorf("kb merge-driver: %w", err)
	}

	fmt.Fprintln(os.Stderr, "kb merge-driver: merge successful")
	return nil
}
