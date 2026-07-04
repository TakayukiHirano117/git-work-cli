package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git-cli/internal/config"
)

type Record struct {
	RepoRoot     string    `json:"repoRoot"`
	ParentBranch string    `json:"parentBranch"`
	ChildBranch  string    `json:"childBranch"`
	IssueKey     string    `json:"issueKey"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Tree struct {
	Records []Record `json:"records"`
}

type Store struct {
	path string
}

func NewDefault() (*Store, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return nil, err
	}
	return New(filepath.Join(dir, "tree.json")), nil
}

func New(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Load() (Tree, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Tree{}, nil
		}
		return Tree{}, err
	}

	var tree Tree
	if err := json.Unmarshal(data, &tree); err != nil {
		return Tree{}, fmt.Errorf("read tree %s: %w", s.path, err)
	}
	return tree, nil
}

func (s *Store) Add(record Record) error {
	tree, err := s.Load()
	if err != nil {
		return err
	}

	for _, existing := range tree.Records {
		if existing.RepoRoot == record.RepoRoot && existing.ChildBranch == record.ChildBranch {
			return DuplicateBranchError(existing.ChildBranch, existing.ParentBranch)
		}
	}

	tree.Records = append(tree.Records, record)
	return s.Save(tree)
}

func (s *Store) Save(tree Tree) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, append(data, '\n'), 0o600)
}

func (t Tree) Children(repoRoot, parentBranch string) []Record {
	records := make([]Record, 0)
	for _, record := range t.Records {
		if record.RepoRoot == repoRoot && record.ParentBranch == parentBranch {
			records = append(records, record)
		}
	}
	return records
}

func (t Tree) ForEpic(repoRoot, epicKey string) []Record {
	records := make([]Record, 0)
	epicPrefix := issuePrefix(epicKey)
	for _, record := range t.Records {
		if record.RepoRoot != repoRoot {
			continue
		}
		if issuePrefix(record.IssueKey) == epicPrefix {
			records = append(records, record)
		}
	}
	return records
}

func DuplicateBranchError(childBranch, parentBranch string) error {
	return fmt.Errorf("branch already recorded: %s (parent: %s)", childBranch, parentBranch)
}

func (t Tree) FindChildBranch(repoRoot, childBranch string) (Record, bool) {
	for _, record := range t.Records {
		if record.RepoRoot == repoRoot && record.ChildBranch == childBranch {
			return record, true
		}
	}
	return Record{}, false
}

func issuePrefix(issueKey string) string {
	upper := strings.ToUpper(issueKey)
	index := strings.LastIndex(upper, "-")
	if index < 0 {
		return upper
	}
	return upper[:index]
}
