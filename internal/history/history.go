package history

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"debugrun/internal/invoke"
)

type Param struct {
	Name   string   `json:"name"`
	Kind   string   `json:"kind"`
	Scalar string   `json:"scalar,omitempty"`
	List   []string `json:"list,omitempty"`
}

type Entry struct {
	Timestamp   time.Time `json:"timestamp"`
	ProfileName string    `json:"profile"`
	Params      []Param   `json:"params"`
	ExtraArgs   []string  `json:"extra_args"`
}

func FromInvocation(inv *invoke.Invocation) Entry {
	params := make([]Param, 0, len(inv.Params))
	for _, bound := range inv.Params {
		param := Param{Name: bound.Spec.Name, Kind: bound.Spec.Kind}
		if bound.Spec.Kind == "string" {
			param.Scalar = bound.Value.Scalar
		} else {
			param.List = append([]string{}, bound.Value.List...)
		}
		params = append(params, param)
	}
	return Entry{
		Timestamp:   time.Now(),
		ProfileName: inv.ProfileName,
		Params:      params,
		ExtraArgs:   append([]string{}, inv.ExtraArgs...),
	}
}

func Append(path string, entry Entry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(entry)
}

func ReadAll(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			return nil, fmt.Errorf("invalid history entry: %w", err)
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func Last(path string) (*Entry, error) {
	entries, err := ReadAll(path)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("history is empty")
	}
	return &entries[len(entries)-1], nil
}

func FromLast(path string, n int) (*Entry, error) {
	entries, err := ReadAll(path)
	if err != nil {
		return nil, err
	}
	if len(entries) < n {
		return nil, fmt.Errorf("history has only %d entries", len(entries))
	}
	return &entries[len(entries)-n], nil
}
