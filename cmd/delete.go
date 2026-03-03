package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"kb/internal/store"
)

// DeleteCmd deletes a document from the knowledge base.
type DeleteCmd struct {
	ID    int64 `arg:"" help:"Document ID to delete."`
	Force bool  `help:"Skip confirmation prompt." short:"f"`
}

// Run executes the delete command.
func (c *DeleteCmd) Run(cli *CLI) error {
	s, err := store.Open(cli.DB)
	if err != nil {
		return err
	}
	defer s.Close()

	if !c.Force {
		doc, err := s.GetDocument(c.ID)
		if err != nil {
			return err
		}
		fmt.Printf("Delete document %d (%q)? [y/N] ", doc.ID, doc.Title)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(answer)) != "y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if err := s.DeleteDocument(c.ID); err != nil {
		return err
	}

	fmt.Printf("Deleted document %d.\n", c.ID)
	return nil
}
