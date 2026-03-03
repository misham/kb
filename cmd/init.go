package cmd

import (
	"fmt"

	"kb/internal/store"
)

// InitCmd creates a new knowledge base database.
type InitCmd struct{}

// Run executes the init command.
func (c *InitCmd) Run(cli *CLI) error {
	s, err := store.Open(cli.DB)
	if err != nil {
		return err
	}
	defer s.Close()

	if err := s.InitSchema(); err != nil {
		return err
	}

	fmt.Printf("Knowledge base created at %s\n", cli.DB)
	return nil
}
