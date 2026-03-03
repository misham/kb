package cmd

import (
	"fmt"
	"strconv"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"kb/internal/store"
)

// ListCmd lists documents in the knowledge base.
type ListCmd struct {
	Type string `help:"Filter by document type." short:"t"`
}

// Run executes the list command.
func (c *ListCmd) Run(cli *CLI) error {
	s, err := store.Open(cli.DB)
	if err != nil {
		return err
	}
	defer s.Close()

	docs, err := s.ListDocuments(c.Type)
	if err != nil {
		return err
	}

	if len(docs) == 0 {
		fmt.Println("No documents found.")
		return nil
	}

	t := table.New().
		Headers("ID", "Title", "Type", "Created").
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, _ int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			return lipgloss.NewStyle()
		})

	for _, d := range docs {
		t.Row(strconv.FormatInt(d.ID, 10), d.Title, d.Type, d.CreatedAt)
	}

	fmt.Println(t)
	return nil
}
