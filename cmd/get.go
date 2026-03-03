package cmd

import (
	"fmt"

	"github.com/charmbracelet/glamour"

	"kb/internal/store"
)

// GetCmd displays a single document.
type GetCmd struct {
	ID int64 `arg:"" help:"Document ID."`
}

// Run executes the get command.
func (c *GetCmd) Run(cli *CLI) error {
	s, err := store.Open(cli.DB)
	if err != nil {
		return err
	}
	defer s.Close()

	doc, err := s.GetDocument(c.ID)
	if err != nil {
		return err
	}

	fmt.Printf("%s  %s\n", headerStyle.Render(doc.Title), faintStyle.Render(fmt.Sprintf("#%d · %s", doc.ID, doc.Type)))
	if len(doc.Metadata) > 0 {
		fmt.Printf("%s %s\n", titleStyle.Render("Meta:"), string(doc.Metadata))
	}
	fmt.Println()

	rendered, err := glamour.Render(doc.Content, "auto")
	if err != nil {
		fmt.Println(doc.Content)
		return nil
	}
	fmt.Print(rendered)
	return nil
}
