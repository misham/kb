package cmd

import (
	"fmt"

	"kb/internal/store"
)

// SearchCmd searches the knowledge base using full-text search.
type SearchCmd struct {
	Query string `arg:"" help:"Search query."`
	Type  string `help:"Filter by document type." short:"t"`
}

// Run executes the search command.
func (c *SearchCmd) Run(cli *CLI) error {
	s, err := store.Open(cli.DB)
	if err != nil {
		return err
	}
	defer s.Close()

	results, err := s.Search(c.Query, c.Type)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	for _, r := range results {
		id := faintStyle.Render(fmt.Sprintf("#%d", r.ID))
		title := titleStyle.Render(r.Title)
		docType := faintStyle.Render(r.Type)
		fmt.Printf("%s %s %s\n    %s\n\n", id, title, docType, r.Snippet)
	}
	return nil
}
