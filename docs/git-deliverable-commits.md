# Git Deliverable Commits

`git-deliverable-commits` creates a commit plan from the current Git working
tree. The goal is a clean history that tells the product story by deliverable,
not a mechanical history grouped by individual files.

## Commands

```bash
go run ./cmd/git-deliverable-commits --dry-run
go run ./cmd/git-deliverable-commits --interactive
go run ./cmd/git-deliverable-commits --interactive --execute
go run ./cmd/git-deliverable-commits --repo /path/to/repo --dry-run
```

By default, the command prints a plan and exits without committing. Use
`--execute` to create commits after reviewing the plan and confirming.

## What It Analyzes

The tool reads:

- staged changes
- unstaged changes
- untracked files
- directory structure
- file names
- Git diff content
- Knowledge Forge deliverable categories

It recognizes common scopes:

- `repo`
- `retrieval`
- `eval`
- `indexing`
- `ui`
- `planning`
- `impact`
- `auth`
- `chat`
- `db`
- `deploy`
- `tooling`

It also separates common commit types:

- `feat(...)`
- `test(...)`
- `docs(...)`
- `refactor(...)`
- `chore(...)`

## Example Output

```text
Deliverable Plan

[1]
feat(repo): add repository service
Files:
- internal/repositories/service.go
- internal/repositories/store.go

[2]
feat(retrieval): add dense retrieval
Files:
- internal/retrieval/dense.go

[3]
test(retrieval): add tests
Files:
- internal/retrieval/retrieval_test.go

[4]
docs(repo): document architecture
Files:
- docs/repository-intelligence.md
```

## Interactive Editing

Run:

```bash
go run ./cmd/git-deliverable-commits --interactive
```

Available commands:

```text
enter                     accept plan
p                         print plan
e <n> <message>           edit commit message
m <file> <n>              move a file to another deliverable
n <message>               create a new deliverable
q                         quit
```

Examples:

```text
e 1 feat(repo): add repository snapshot model
m internal/retrieval/dense.go 2
n docs(repo): document repository architecture
```

After creating a new deliverable, move files into it with `m <file> <n>`.

## Execution Rules

When `--execute` is used, the tool:

1. Prints the proposed plan.
2. Lets you edit the plan when `--interactive` is enabled.
3. Asks for confirmation.
4. Stages files one deliverable at a time.
5. Creates one commit per non-empty deliverable.
6. Skips any deliverable that has no staged diff.

The tool never:

- pushes automatically
- amends existing commits
- rebases or rewrites history
- creates empty commits

## Mixed Staged and Unstaged Files

If a file has both staged and unstaged changes, the plan shows a warning. During
execution, the tool commits the full current file content for that deliverable.
If you need hunk-level precision, split the file manually before running the
tool.

## Design Intent

The success metric is not more commits. The success metric is a reviewer being
able to understand project evolution by reading commit history alone.
