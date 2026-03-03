package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adrg/frontmatter"

	"kb/internal/store"
)

var (
	headingRe    = regexp.MustCompile(`(?m)^#\s+(.+)$`)
	datePrefixRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-`)
)

// ImportCmd imports a markdown file into the knowledge base.
type ImportCmd struct {
	File string `arg:"" help:"Markdown file to import." type:"existingfile"`
	Type string `required:"" short:"t" help:"Document type (e.g., research, plan)."`
}

// Run executes the import command.
func (c *ImportCmd) Run(cli *CLI) error {
	data, err := os.ReadFile(c.File)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	var meta map[string]interface{}
	content, err := frontmatter.Parse(strings.NewReader(string(data)), &meta)
	if err != nil {
		// If frontmatter parsing fails, use the whole file as content.
		content = data
		meta = nil
	}

	title := extractTitle(meta, string(content), filepath.Base(c.File))

	var metadata json.RawMessage
	if len(meta) > 0 {
		metadata, err = json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("marshaling metadata: %w", err)
		}
	}

	s, err := store.Open(cli.DB)
	if err != nil {
		return err
	}
	defer s.Close()

	id, err := s.InsertDocument(c.Type, title, string(content), metadata)
	if err != nil {
		return err
	}

	fmt.Printf("Imported document %d: %s\n", id, title)
	return nil
}

// extractTitle determines the document title from metadata, content headings, or filename.
func extractTitle(meta map[string]interface{}, content string, filename string) string {
	if topic, ok := meta["topic"].(string); ok && topic != "" {
		return topic
	}
	if title, ok := meta["title"].(string); ok && title != "" {
		return title
	}
	if m := headingRe.FindStringSubmatch(content); m != nil {
		return m[1]
	}
	// Fall back to filename: strip extension and date prefix.
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	// Strip leading date prefix like "2024-01-01-".
	return datePrefixRe.ReplaceAllString(name, "")
}
