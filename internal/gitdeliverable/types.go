package gitdeliverable

import (
	"fmt"
	"strings"
)

type Change struct {
	Path        string
	OrigPath    string
	Status      string
	Staged      bool
	Unstaged    bool
	Untracked   bool
	Deleted     bool
	Renamed     bool
	Diff        string
	ContentHint string
}

type Deliverable struct {
	Message string
	Type    string
	Scope   string
	Topic   string
	Files   []Change
}

type Plan struct {
	Deliverables []Deliverable
	Warnings     []string
}

func (p Plan) Empty() bool {
	return len(p.Deliverables) == 0
}

func (d Deliverable) FilePaths() []string {
	paths := make([]string, 0, len(d.Files))
	seen := map[string]bool{}
	for _, file := range d.Files {
		for _, path := range file.StagePaths() {
			if path == "" || seen[path] {
				continue
			}
			seen[path] = true
			paths = append(paths, path)
		}
	}
	return paths
}

func (c Change) DisplayPath() string {
	if c.Path != "" {
		return c.Path
	}
	return c.OrigPath
}

func (c Change) StagePaths() []string {
	if c.Renamed && c.OrigPath != "" && c.Path != "" {
		return []string{c.OrigPath, c.Path}
	}
	return []string{c.DisplayPath()}
}

func (p *Plan) Normalize() {
	out := p.Deliverables[:0]
	for _, deliverable := range p.Deliverables {
		if len(deliverable.Files) == 0 {
			continue
		}
		deliverable.Message = strings.TrimSpace(deliverable.Message)
		if deliverable.Message == "" {
			deliverable.Message = defaultMessage(deliverable.Type, deliverable.Scope)
		}
		out = append(out, deliverable)
	}
	p.Deliverables = out
}

func (p Plan) FindFile(query string) (int, int, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return -1, -1, fmt.Errorf("file query is empty")
	}
	var matchDeliverable, matchFile int
	matches := 0
	for i, deliverable := range p.Deliverables {
		for j, file := range deliverable.Files {
			path := file.DisplayPath()
			if path == query || strings.HasSuffix(path, "/"+query) || baseName(path) == query {
				matchDeliverable = i
				matchFile = j
				matches++
			}
		}
	}
	if matches == 0 {
		return -1, -1, fmt.Errorf("file %q not found in plan", query)
	}
	if matches > 1 {
		return -1, -1, fmt.Errorf("file %q is ambiguous; use a full path", query)
	}
	return matchDeliverable, matchFile, nil
}

func (p *Plan) MoveFile(fileQuery string, targetNumber int) error {
	if targetNumber < 1 || targetNumber > len(p.Deliverables) {
		return fmt.Errorf("target deliverable must be between 1 and %d", len(p.Deliverables))
	}
	sourceIdx, fileIdx, err := p.FindFile(fileQuery)
	if err != nil {
		return err
	}
	targetIdx := targetNumber - 1
	if sourceIdx == targetIdx {
		return nil
	}
	file := p.Deliverables[sourceIdx].Files[fileIdx]
	p.Deliverables[sourceIdx].Files = append(p.Deliverables[sourceIdx].Files[:fileIdx], p.Deliverables[sourceIdx].Files[fileIdx+1:]...)
	p.Deliverables[targetIdx].Files = append(p.Deliverables[targetIdx].Files, file)
	p.Normalize()
	return nil
}

func (p *Plan) EditMessage(number int, message string) error {
	if number < 1 || number > len(p.Deliverables) {
		return fmt.Errorf("deliverable must be between 1 and %d", len(p.Deliverables))
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}
	p.Deliverables[number-1].Message = message
	return nil
}

func (p *Plan) AddDeliverable(message string) error {
	message = strings.TrimSpace(message)
	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}
	typ, scope := parseConventionalPrefix(message)
	p.Deliverables = append(p.Deliverables, Deliverable{
		Message: message,
		Type:    typ,
		Scope:   scope,
		Topic:   slugWords(message),
	})
	return nil
}

func defaultMessage(typ, scope string) string {
	if typ == "" {
		typ = "chore"
	}
	if scope == "" {
		return typ + ": update deliverable"
	}
	return fmt.Sprintf("%s(%s): update deliverable", typ, scope)
}
