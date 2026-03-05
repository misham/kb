package cmd

import (
	"encoding/json"
	"errors"
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

// ImportCmd imports markdown files into the knowledge base.
type ImportCmd struct {
	Files []string `arg:"" help:"One or more markdown files to import." type:"existingfile"`
	Type  string   `required:"" short:"t" help:"Document type (e.g., research, plan)."`
}

// Run executes the import command.
func (c *ImportCmd) Run(cli *CLI) error {
	s, err := store.Open(cli.DB)
	if err != nil {
		return err
	}
	defer s.Close()

	var errs []error
	for _, path := range c.Files {
		id, title, err := c.importFile(s, path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error importing %s: %v\n", path, err)
			errs = append(errs, fmt.Errorf("%s: %w", path, err))
			continue
		}
		fmt.Printf("Imported document %d: %s\n", id, title)
	}

	total := len(c.Files)
	failed := len(errs)
	if total > 1 || failed > 0 {
		fmt.Printf("Imported %d of %d files (%d failed)\n", total-failed, total, failed)
	}
	return errors.Join(errs...)
}

// importFile reads, parses, and inserts a single markdown file.
func (c *ImportCmd) importFile(s *store.Store, path string) (int64, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, "", fmt.Errorf("reading file: %w", err)
	}

	var meta map[string]interface{}
	content, err := frontmatter.Parse(strings.NewReader(string(data)), &meta)
	if err != nil {
		content = data
		meta = nil
	}

	title := extractTitle(meta, string(content), filepath.Base(path))

	var metadata json.RawMessage
	if len(meta) > 0 {
		metadata, err = json.Marshal(meta)
		if err != nil {
			return 0, "", fmt.Errorf("marshaling metadata: %w", err)
		}
	}

	id, err := s.InsertDocument(c.Type, title, string(content), metadata)
	if err != nil {
		return 0, "", err
	}

	return id, title, nil
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
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	return datePrefixRe.ReplaceAllString(name, "")
}
