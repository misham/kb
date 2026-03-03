package cmd

import (
	"fmt"

	"kb/internal/store"
)

// LinkCmd creates a link between two documents.
type LinkCmd struct {
	ID1 int64  `arg:"" help:"First document ID."`
	ID2 int64  `arg:"" help:"Second document ID."`
	Rel string `help:"Relationship type." short:"r"`
}

// Run executes the link command.
func (c *LinkCmd) Run(cli *CLI) error {
	s, err := store.Open(cli.DB)
	if err != nil {
		return err
	}
	defer s.Close()

	if err := s.CreateLink(c.ID1, c.ID2, c.Rel); err != nil {
		return err
	}

	fmt.Printf("Linked document %d → %d", c.ID1, c.ID2)
	if c.Rel != "" {
		fmt.Printf(" (%s)", c.Rel)
	}
	fmt.Println(".")
	return nil
}
