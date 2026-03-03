package cmd

import (
	"fmt"

	"kb/internal/store"
)

// LinksCmd shows documents linked to a given document.
type LinksCmd struct {
	ID int64 `arg:"" help:"Document ID."`
}

// Run executes the links command.
func (c *LinksCmd) Run(cli *CLI) error {
	s, err := store.Open(cli.DB)
	if err != nil {
		return err
	}
	defer s.Close()

	links, err := s.GetLinks(c.ID)
	if err != nil {
		return err
	}

	if len(links) == 0 {
		fmt.Println("No linked documents.")
		return nil
	}

	for _, l := range links {
		id := faintStyle.Render(fmt.Sprintf("#%d", l.ID))
		title := titleStyle.Render(l.Title)
		dir := faintStyle.Render(l.Direction)
		rel := ""
		if l.Relationship != "" {
			rel = " " + headerStyle.Render("["+l.Relationship+"]")
		}
		fmt.Printf("%s %s %s%s — %s\n", id, title, dir, rel, l.Type)
	}
	return nil
}
