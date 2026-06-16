package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/tarunngusain08/knowledge-forge/internal/gitdeliverable"
)

func main() {
	var (
		dryRun      = flag.Bool("dry-run", false, "print the deliverable commit plan without creating commits")
		interactive = flag.Bool("interactive", false, "edit commit messages and move files before accepting the plan")
		execute     = flag.Bool("execute", false, "create commits after printing the plan and receiving confirmation")
		repoDir     = flag.String("repo", ".", "git repository directory to analyze")
	)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: git-deliverable-commits [--dry-run] [--interactive] [--execute] [--repo <path>]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *dryRun && *execute {
		exitError("--dry-run and --execute cannot be used together")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	git := gitdeliverable.Git{Dir: *repoDir}
	changes, err := git.Changes(ctx)
	if err != nil {
		exitError(err.Error())
	}
	plan := gitdeliverable.BuildPlan(changes)
	if plan.Empty() {
		gitdeliverable.PrintPlan(os.Stdout, plan)
		return
	}

	accepted := true
	if *interactive {
		plan, accepted, err = gitdeliverable.EditPlan(os.Stdin, os.Stdout, plan)
		if err != nil {
			exitError(err.Error())
		}
		if !accepted {
			fmt.Fprintln(os.Stdout, "Aborted.")
			return
		}
	} else {
		gitdeliverable.PrintPlan(os.Stdout, plan)
	}

	if !*execute {
		fmt.Fprintln(os.Stdout, "Dry run only. Re-run with --execute to create commits.")
		return
	}

	confirmed, err := gitdeliverable.Confirm(os.Stdin, os.Stdout, "Create these commits now?")
	if err != nil {
		exitError(err.Error())
	}
	if !confirmed {
		fmt.Fprintln(os.Stdout, "Aborted.")
		return
	}

	results, err := git.ExecutePlan(ctx, plan)
	if err != nil {
		exitError(err.Error())
	}
	gitdeliverable.PrintCommitResults(os.Stdout, results)
}

func exitError(message string) {
	fmt.Fprintln(os.Stderr, "error:", message)
	os.Exit(1)
}
