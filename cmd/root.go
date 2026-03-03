package cmd

// CLI defines the top-level command structure for kb.
type CLI struct {
	DB     string    `help:"Path to database." default:"kb.db" env:"KB_DB"`
	Plain  bool      `help:"Disable styled output."`
	Init   InitCmd   `cmd:"" help:"Create a new knowledge base."`
	Import ImportCmd `cmd:"" help:"Import a markdown file."`
	Search SearchCmd `cmd:"" help:"Search the knowledge base."`
	List   ListCmd   `cmd:"" help:"List documents."`
	Get    GetCmd    `cmd:"" help:"Display a document."`
	Delete DeleteCmd `cmd:"" help:"Delete a document."`
	Link   LinkCmd   `cmd:"" help:"Link two documents."`
	Links  LinksCmd  `cmd:"" help:"Show linked documents."`
}
