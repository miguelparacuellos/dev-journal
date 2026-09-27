// Package journal is the public application boundary shared by the CLI and TUI.
package journal

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

type Entry struct {
	ID      string `json:"id"`
	Workday string `json:"workday"`
	Text    string `json:"text"`
}
type Data struct {
	Version int     `json:"version"`
	Entries []Entry `json:"entries"`
}
type Journal struct {
	path string
	data Data
}

func Open(path string) (*Journal, error) { j := &Journal{path: path}; err := j.reload(); return j, err }
func (j *Journal) reload() error {
	j.data = Data{Version: 1, Entries: []Entry{}}
	b, err := os.ReadFile(j.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, &j.data); err != nil {
		return fmt.Errorf("journal cannot be read: %w", err)
	}
	if j.data.Version != 1 {
		return fmt.Errorf("unsupported journal version: %d", j.data.Version)
	}
	return nil
}
func (j *Journal) Entries(day string) []Entry {
	result := []Entry{}
	for _, e := range j.data.Entries {
		if e.Workday == day {
			result = append(result, e)
		}
	}
	return result
}
func (j *Journal) Workdays() []string {
	set := map[string]bool{}
	for _, e := range j.data.Entries {
		set[e.Workday] = true
	}
	days := []string{}
	for day := range set {
		days = append(days, day)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(days)))
	return days
}
func validate(day, text string) error {
	if _, err := time.Parse("2006-01-02", day); err != nil {
		return errors.New("workday must be YYYY-MM-DD")
	}
	if strings.TrimSpace(text) == "" {
		return errors.New("entry cannot be empty")
	}
	return nil
}
func (j *Journal) Capture(day, text string) (Entry, error) {
	if err := validate(day, text); err != nil {
		return Entry{}, err
	}
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return Entry{}, err
	}
	e := Entry{hex.EncodeToString(id), day, text}
	err := j.change(func() error { j.data.Entries = append(j.data.Entries, e); return nil })
	return e, err
}
func (j *Journal) Correct(id, text string) error {
	if strings.TrimSpace(text) == "" {
		return errors.New("entry cannot be empty")
	}
	return j.change(func() error {
		for i := range j.data.Entries {
			if j.data.Entries[i].ID == id {
				j.data.Entries[i].Text = text
				return nil
			}
		}
		return errors.New("entry not found")
	})
}
func (j *Journal) change(update func() error) error {
	if err := os.MkdirAll(filepath.Dir(j.path), 0700); err != nil {
		return err
	}
	lock, err := os.OpenFile(j.path+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	if err = j.reload(); err != nil {
		return err
	}
	if err = update(); err != nil {
		return err
	}
	b, err := json.MarshalIndent(j.data, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(j.path), ".journal-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.Write(b); err == nil {
		err = tmp.Sync()
	}
	closeErr := tmp.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(tmp.Name(), j.path)
	}
	if err != nil {
		_ = j.reload()
		return err
	}
	return nil
}
