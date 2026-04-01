package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/jaeyeom/gh-repox/internal/config"
	"github.com/jaeyeom/gh-repox/internal/diff"
)

// CreateResult holds the result of a create operation.
type CreateResult struct {
	Command     string         `json:"command"`
	Repo        string         `json:"repo"`
	URL         string         `json:"url"`
	Created     bool           `json:"created"`
	OwnerSource string         `json:"owner_source"`
	Applied     map[string]any `json:"applied"`
	Clone       CloneResult    `json:"clone"`
	Warnings    []string       `json:"warnings"`
}

// CloneResult holds clone operation results.
type CloneResult struct {
	Requested bool   `json:"requested"`
	Completed bool   `json:"completed"`
	Directory string `json:"directory,omitempty"`
}

// DiffResult holds the result of a diff operation.
type DiffResult struct {
	Command     string       `json:"command"`
	Repo        string       `json:"repo"`
	Differences []diff.Entry `json:"differences"`
}

// ApplyResult holds the result of an apply operation.
type ApplyResult struct {
	Command  string   `json:"command"`
	Repo     string   `json:"repo"`
	Applied  []string `json:"applied"`
	Warnings []string `json:"warnings"`
}

// PrintCreateHuman prints the create result in human-readable format.
func PrintCreateHuman(w io.Writer, r *CreateResult) {
	fmt.Fprintf(w, "Created repository: %s\n", r.URL)
	if len(r.Applied) > 0 {
		fmt.Fprintln(w, "\nApplied settings:")
		for k, v := range r.Applied {
			fmt.Fprintf(w, "- %s: %v\n", k, v)
		}
	}
	fmt.Fprintln(w, "\nClone:")
	if !r.Clone.Requested {
		fmt.Fprintln(w, "- skipped")
	} else if r.Clone.Completed {
		fmt.Fprintf(w, "- completed at %s\n", r.Clone.Directory)
	} else {
		fmt.Fprintln(w, "- failed")
	}
	if len(r.Warnings) > 0 {
		fmt.Fprintln(w, "\nWarnings:")
		for _, warn := range r.Warnings {
			fmt.Fprintf(w, "- %s\n", warn)
		}
	}
}

// PrintApplyHuman prints the apply result in human-readable format.
func PrintApplyHuman(w io.Writer, r *ApplyResult) {
	fmt.Fprintf(w, "Applied policy to: %s\n", r.Repo)
	if len(r.Applied) > 0 {
		fmt.Fprintln(w, "\nApplied:")
		for _, a := range r.Applied {
			fmt.Fprintf(w, "- %s\n", a)
		}
	}
	if len(r.Warnings) > 0 {
		fmt.Fprintln(w, "\nWarnings:")
		for _, warn := range r.Warnings {
			fmt.Fprintf(w, "- %s\n", warn)
		}
	}
}

// PrintDiffHuman prints the diff result in human-readable format.
func PrintDiffHuman(w io.Writer, r *DiffResult) {
	fmt.Fprintf(w, "Diff for: %s\n\n", r.Repo)
	fmt.Fprint(w, diff.FormatHuman(r.Differences))
}

// PrintJSON prints any result as JSON.
func PrintJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// PrintConfigShow prints resolved config values.
func PrintConfigShow(w io.Writer, entries []config.Entry) {
	for _, e := range entries {
		fmt.Fprintf(w, "%-30s %v\n", e.Key, e.Value)
	}
}

// PrintConfigExplain prints resolved config values with sources.
func PrintConfigExplain(w io.Writer, entries []config.Entry) {
	for _, e := range entries {
		fmt.Fprintf(w, "%-30s %-20v source=%s\n", e.Key, e.Value, e.Source)
	}
}

// PrintDryRun prints planned commands.
func PrintDryRun(w io.Writer, header string, commands []string) {
	fmt.Fprintln(w, header)
	fmt.Fprintln(w, "\nPlanned commands:")
	for i, cmd := range commands {
		fmt.Fprintf(w, "%d. %s\n", i+1, cmd)
	}
}

// FormatCommand formats a command and args for display.
func FormatCommand(name string, args ...string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, name)
	parts = append(parts, args...)
	return strings.Join(parts, " ")
}
