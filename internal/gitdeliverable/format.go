package gitdeliverable

import (
	"fmt"
	"io"
	"path/filepath"
)

func PrintPlan(w io.Writer, plan Plan) {
	fmt.Fprintln(w, "Deliverable Plan")
	fmt.Fprintln(w)
	if len(plan.Warnings) > 0 {
		fmt.Fprintln(w, "Warnings:")
		for _, warning := range plan.Warnings {
			fmt.Fprintf(w, "- %s\n", warning)
		}
		fmt.Fprintln(w)
	}
	if len(plan.Deliverables) == 0 {
		fmt.Fprintln(w, "No staged or unstaged changes found.")
		return
	}
	for i, deliverable := range plan.Deliverables {
		fmt.Fprintf(w, "[%d]\n", i+1)
		fmt.Fprintln(w, deliverable.Message)
		fmt.Fprintln(w, "Files:")
		for _, file := range deliverable.Files {
			display := file.DisplayPath()
			if display == "" {
				display = filepath.Base(file.OrigPath)
			}
			fmt.Fprintf(w, "- %s%s\n", display, statusLabel(file))
		}
		fmt.Fprintln(w)
	}
}

func PrintCommitResults(w io.Writer, results []CommitResult) {
	if len(results) == 0 {
		return
	}
	fmt.Fprintln(w, "Commit Results")
	fmt.Fprintln(w)
	for _, result := range results {
		if result.Skipped {
			fmt.Fprintf(w, "- skipped: %s (%s)\n", result.Message, result.Reason)
			continue
		}
		fmt.Fprintf(w, "- %s %s\n", result.Hash, result.Message)
	}
}

func statusLabel(change Change) string {
	switch {
	case change.Untracked:
		return " (untracked)"
	case change.Renamed:
		return " (renamed)"
	case change.Deleted:
		return " (deleted)"
	case change.Staged && change.Unstaged:
		return " (staged + unstaged)"
	case change.Staged:
		return " (staged)"
	case change.Unstaged:
		return " (unstaged)"
	default:
		return ""
	}
}
