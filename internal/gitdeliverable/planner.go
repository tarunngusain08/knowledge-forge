package gitdeliverable

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type groupKey struct {
	Type  string
	Scope string
	Topic string
}

func BuildPlan(changes []Change) Plan {
	groups := map[groupKey][]Change{}
	order := make([]groupKey, 0)
	warnings := mixedStateWarnings(changes)
	for _, change := range changes {
		key := classify(change)
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		groups[key] = append(groups[key], change)
	}
	sort.SliceStable(order, func(i, j int) bool {
		return groupSortRank(order[i]) < groupSortRank(order[j])
	})
	deliverables := make([]Deliverable, 0, len(order))
	for _, key := range order {
		files := groups[key]
		sort.SliceStable(files, func(i, j int) bool {
			return files[i].DisplayPath() < files[j].DisplayPath()
		})
		deliverables = append(deliverables, Deliverable{
			Message: messageFor(key, files),
			Type:    key.Type,
			Scope:   key.Scope,
			Topic:   key.Topic,
			Files:   files,
		})
	}
	return Plan{Deliverables: deliverables, Warnings: warnings}
}

func classify(change Change) groupKey {
	path := strings.ToLower(change.DisplayPath())
	diff := strings.ToLower(change.Diff)
	signals := strings.ToLower(change.Diff + "\n" + change.ContentHint)
	scope := detectScope(path, signals)
	typ := detectType(path, diff)
	topic := detectTopic(path, signals, typ, scope)
	return groupKey{Type: typ, Scope: scope, Topic: topic}
}

func detectType(path, diff string) string {
	switch {
	case isDocPath(path):
		return "docs"
	case isTestPath(path):
		return "test"
	case isChorePath(path):
		return "chore"
	case strings.Contains(diff, "rename from") && strings.Contains(diff, "rename to"):
		return "refactor"
	default:
		return "feat"
	}
}

func detectScope(path, diff string) string {
	switch {
	case containsAny(path, "cmd/git-deliverable-commits", "internal/gitdeliverable", "git-deliverable-commits") ||
		containsAny(addedOnly(diff), "git-deliverable-commits", "deliverable commit"):
		return "tooling"
	case containsAny(path, "internal/codeqa", "planning", "/plans", "plan"):
		return "planning"
	case containsAny(path, "impact"):
		return "impact"
	case containsAny(path, "internal/repositories", "repository", "repositories", "codeintel", "snapshot"):
		return "repo"
	case containsAny(path, "internal/retrieval", "retrieval", "dense", "lexical", "rerank", "rrf", "fts"):
		return "retrieval"
	case containsAny(path, "internal/evaluation", "eval-runner", "benchmark", "ragas", "evaluation"):
		return "eval"
	case containsAny(path, "internal/indexing", "indexer", "worker", "chunker", "extractor", "ingestion"):
		return "indexing"
	case strings.HasPrefix(path, "ui/") || containsAny(path, "react", "vite", "streamlit", "demo mode"):
		return "ui"
	case containsAny(path, "internal/auth", "jwt", "password"):
		return "auth"
	case containsAny(path, "internal/chat", "chat"):
		return "chat"
	case containsAny(path, "internal/costs", "cost"):
		return "costs"
	case containsAny(path, "migrations/", "queries/", "internal/db", "database"):
		return "db"
	case containsAny(path, "deploy/", "docker", "cloud-run"):
		return "deploy"
	case containsAny(diff, "repository", "snapshot"):
		return "repo"
	case containsAny(diff, "retrieval", "pinecone", "fts", "rerank"):
		return "retrieval"
	case containsAny(diff, "benchmark", "metric", "ragas"):
		return "eval"
	default:
		return "project"
	}
}

func detectTopic(path, diff, typ, scope string) string {
	if typ == "docs" {
		if scope == "repo" || containsAny(path, "architecture", "repository") {
			return "architecture"
		}
		if containsAny(path, "evaluation", "benchmark") {
			return "evaluation"
		}
		if scope == "tooling" {
			return "deliverable commit tool"
		}
		return "docs"
	}
	if typ == "test" {
		return "tests"
	}
	switch scope {
	case "repo":
		if containsAny(path, "snapshot", "branch", "codeintel") || containsAny(diff, "snapshot", "branch") {
			return "snapshot model"
		}
		if containsAny(path, "ingestion", "indexer") || containsAny(diff, "ingestion", "index") {
			return "ingestion service"
		}
		return "repository service"
	case "retrieval":
		if containsAny(path, "dense") || containsAny(diff, "dense", "pinecone") {
			return "dense retrieval"
		}
		if containsAny(path, "fts", "lexical") || containsAny(diff, "lexical", "full text", "fts") {
			return "lexical retrieval"
		}
		if containsAny(path, "rrf", "fusion") || containsAny(diff, "fusion", "reciprocal rank") {
			return "retrieval fusion"
		}
		return "repository retrieval"
	case "eval":
		if containsAny(path, "fixture", "benchmarks") || containsAny(diff, "fixture", "corpus") {
			return "benchmark corpus"
		}
		if containsAny(diff, "threshold", "complexity", "gate") {
			return "complexity review gates"
		}
		return "evaluation metrics"
	case "indexing":
		return "indexing worker"
	case "ui":
		if containsAny(path, "demo", "app.tsx") || containsAny(diff, "demo mode") {
			return "demo mode"
		}
		return "product ui"
	case "planning":
		return "grounded planning workflow"
	case "impact":
		return "impact analysis workflow"
	case "tooling":
		return "deliverable commit tool"
	default:
		return scope
	}
}

func messageFor(key groupKey, files []Change) string {
	switch key.Type {
	case "docs":
		return conventional(key.Type, key.Scope, "document "+nounPhrase(key))
	case "test":
		return conventional(key.Type, key.Scope, "add "+nounPhrase(key))
	case "refactor":
		return conventional(key.Type, key.Scope, "organize "+nounPhrase(key))
	case "chore":
		return conventional(key.Type, key.Scope, "update "+nounPhrase(key))
	default:
		return conventional(key.Type, key.Scope, "add "+nounPhrase(key))
	}
}

func nounPhrase(key groupKey) string {
	if key.Topic == "" || key.Topic == key.Scope {
		switch key.Scope {
		case "repo":
			return "repository changes"
		case "retrieval":
			return "retrieval changes"
		case "eval":
			return "evaluation changes"
		case "ui":
			return "ui changes"
		case "tooling":
			return "tooling changes"
		default:
			return key.Scope + " changes"
		}
	}
	return key.Topic
}

func conventional(typ, scope, summary string) string {
	scope = strings.TrimSpace(scope)
	summary = strings.TrimSpace(summary)
	if scope == "" || scope == "project" {
		return typ + ": " + summary
	}
	return typ + "(" + scope + "): " + summary
}

func isDocPath(path string) bool {
	return strings.HasPrefix(path, "docs/") ||
		strings.HasPrefix(path, "deploy/") ||
		path == "readme.md" ||
		strings.HasSuffix(path, ".md")
}

func isTestPath(path string) bool {
	base := filepath.Base(path)
	return strings.Contains(path, "/tests/") ||
		strings.Contains(path, "/test/") ||
		strings.HasSuffix(base, "_test.go") ||
		strings.HasSuffix(base, ".test.ts") ||
		strings.HasSuffix(base, ".spec.ts") ||
		strings.HasPrefix(base, "test_")
}

func isChorePath(path string) bool {
	base := filepath.Base(path)
	return base == "go.mod" ||
		base == "go.sum" ||
		base == "package-lock.json" ||
		base == "makefile" ||
		strings.HasPrefix(base, "dockerfile") ||
		path == "docker-compose.yml" ||
		strings.HasPrefix(path, ".github/") ||
		strings.HasPrefix(path, "migrations/")
}

func groupSortRank(key groupKey) int {
	typeRank := map[string]int{
		"feat":     10,
		"refactor": 20,
		"test":     30,
		"docs":     40,
		"chore":    50,
	}
	scopeRank := map[string]int{
		"repo":      1,
		"indexing":  2,
		"retrieval": 3,
		"planning":  4,
		"impact":    5,
		"eval":      6,
		"ui":        7,
		"tooling":   8,
	}
	return typeRank[key.Type]*100 + scopeRank[key.Scope]
}

func mixedStateWarnings(changes []Change) []string {
	var warnings []string
	for _, change := range changes {
		if change.Staged && change.Unstaged {
			warnings = append(warnings, "file has both staged and unstaged changes; execution will commit the full current file: "+change.DisplayPath())
		}
	}
	return warnings
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func baseName(path string) string {
	return filepath.Base(path)
}

func parseConventionalPrefix(message string) (string, string) {
	pattern := regexp.MustCompile(`^([a-z]+)(?:\(([^)]+)\))?:`)
	matches := pattern.FindStringSubmatch(strings.TrimSpace(message))
	if len(matches) == 0 {
		return "chore", "project"
	}
	return matches[1], matches[2]
}

func slugWords(message string) string {
	message = regexp.MustCompile(`^[a-z]+(?:\([^)]+\))?:\s*`).ReplaceAllString(strings.ToLower(message), "")
	words := strings.Fields(message)
	if len(words) > 4 {
		words = words[:4]
	}
	return strings.Join(words, " ")
}

func addedOnly(diff string) string {
	var b strings.Builder
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+++") || !strings.HasPrefix(line, "+") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}
