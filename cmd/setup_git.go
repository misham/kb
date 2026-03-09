package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// SetupGitCmd configures the local git repository to use kb as a merge driver
// for *.db files, so that SQLite database merges are handled automatically.
type SetupGitCmd struct{}

// Run configures the git merge driver in the local repo.
func (c *SetupGitCmd) Run(_ *CLI) error {
	commands := []struct {
		args []string
		desc string
	}{
		// Merge driver: runs automatically during git merge.
		{
			args: []string{"git", "config", "merge.kb.name", "kb database merge driver"},
			desc: "setting merge driver name",
		},
		{
			args: []string{"git", "config", "merge.kb.driver", "kb merge-driver %O %A %B"},
			desc: "setting merge driver command",
		},
		// Mergetool: runs manually via `git mergetool` when a conflict
		// was not resolved by the merge driver. We copy LOCAL to MERGED
		// first so merge-driver can modify it in place without touching
		// the original LOCAL file.
		{
			args: []string{
				"git", "config", "mergetool.kb.cmd",
				`cp "$LOCAL" "$MERGED" && kb merge-driver "$BASE" "$MERGED" "$REMOTE"`,
			},
			desc: "setting mergetool command",
		},
		{
			args: []string{"git", "config", "mergetool.kb.trustExitCode", "true"},
			desc: "setting mergetool trustExitCode",
		},
	}

	for _, cmd := range commands {
		out, err := exec.CommandContext(context.Background(), cmd.args[0], cmd.args[1:]...).CombinedOutput() //nolint:gosec // args are hardcoded literals
		if err != nil {
			return fmt.Errorf("%s: %w\n%s", cmd.desc, err, out)
		}
	}

	fmt.Println("Git merge driver configured.")
	fmt.Println("")
	fmt.Println("Added to .git/config:")
	fmt.Println("  [merge \"kb\"]")
	fmt.Println("    name = kb database merge driver")
	fmt.Println("    driver = kb merge-driver %O %A %B")
	fmt.Println("")
	fmt.Println("  [mergetool \"kb\"]")
	fmt.Println(`    cmd = cp "$LOCAL" "$MERGED" && kb merge-driver "$BASE" "$MERGED" "$REMOTE"`)
	fmt.Println("    trustExitCode = true")
	fmt.Println("")
	if !gitattributesHasKBMerge() {
		fmt.Println("WARNING: .gitattributes does not contain '*.db merge=kb'.")
		fmt.Println("The merge driver will NOT activate without this line.")
		fmt.Println("Add it with:")
		fmt.Println("  echo '*.db merge=kb' >> .gitattributes")
	} else {
		fmt.Println(".gitattributes already contains '*.db merge=kb'.")
	}
	fmt.Println("")
	fmt.Println("If the merge driver can't resolve a conflict, run:")
	fmt.Println("  git mergetool --tool=kb")
	return nil
}

// gitattributesHasKBMerge checks whether git is configured to use the "kb"
// merge driver for *.db files. It first tries `git check-attr` (which respects
// all .gitattributes locations), then falls back to reading .gitattributes
// directly if git is unavailable.
func gitattributesHasKBMerge() bool {
	out, err := exec.CommandContext(context.Background(), "git", "check-attr", "merge", "--", "test.db").Output() //nolint:gosec // args are hardcoded literals
	if err == nil {
		return strings.Contains(string(out), "merge: kb")
	}
	// Fallback: check the repo-root .gitattributes directly.
	data, err := os.ReadFile(".gitattributes")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "*.db merge=kb" {
			return true
		}
	}
	return false
}
