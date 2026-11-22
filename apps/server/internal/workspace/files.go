package workspace

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider"
)

// Keep ordinary entries visible even when the owner creates a file or directory named Attachments.
func attachmentDirectory(d provider.Filesystem) (string, error) {
	entries, err := fs.ReadDir(d, ".")
	if err != nil {
		return "", err
	}
	taken := make(map[string]bool, len(entries))
	for _, entry := range entries {
		taken[entry.Name()] = true
	}
	name := "Attachments"
	for n := 2; taken[name]; n++ {
		name = fmt.Sprintf("Attachments (%d)", n)
	}
	return name, nil
}

func (s *Service) entries(ctx context.Context, w, dir string, d provider.Filesystem) ([]Entry, error) {
	if !fs.ValidPath(dir) {
		return nil, fs.ErrInvalid
	}
	byName := map[string]Entry{}
	managed, err := attachmentDirectory(d)
	if err != nil {
		return nil, err
	}
	var entries []fs.DirEntry
	if dir != managed {
		entries, err = fs.ReadDir(d, dir)
		if err != nil {
			return nil, err
		}
	}
	for _, e := range entries {
		kind := "file"
		if e.IsDir() {
			kind = "directory"
		}
		byName[e.Name()] = Entry{path.Join(dir, e.Name()), e.Name(), kind}
	}
	if dir == "." || dir == managed {
		attachments, err := s.listAttachments(ctx, w)
		if err != nil {
			return nil, err
		}
		for _, attachment := range attachments {
			if _, err := d.ReadVersion(attachment.Key); err != nil {
				return nil, err
			}
			if dir == "." {
				byName[managed] = Entry{managed, managed, "directory"}
			} else {
				byName[attachment.Name] = Entry{path.Join(dir, attachment.Name), attachment.Name, "file"}
			}
		}
	}
	if dir == managed && len(byName) == 0 {
		return nil, fs.ErrNotExist
	}
	result := make([]Entry, 0, len(byName))
	for _, e := range byName {
		result = append(result, e)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func (s *Service) Files(ctx context.Context, w, dir, search string) ([]Entry, error) {
	if dir == "" {
		dir = "."
	}
	d, err := s.open(ctx, w)
	if err != nil {
		return nil, err
	}
	defer d.Close()
	if search == "" {
		return s.entries(ctx, w, dir, d)
	}
	var visit func(string) error
	found := []Entry{}
	needle := strings.ToLower(search)
	visit = func(dir string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		entries, err := s.entries(ctx, w, dir, d)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if strings.Contains(strings.ToLower(e.Name), needle) {
				found = append(found, e)
			}
			if e.Kind == "directory" {
				if err := visit(e.Path); err != nil {
					return err
				}
			}
		}
		return nil
	}
	err = visit(dir)
	return found, err
}

func (s *Service) Read(ctx context.Context, w, name string) ([]byte, error) {
	if !fs.ValidPath(name) || name == "." {
		return nil, fs.ErrInvalid
	}
	d, err := s.open(ctx, w)
	if err != nil {
		return nil, err
	}
	defer d.Close()
	if id, ok, err := s.attachment(ctx, w, name, d); err != nil {
		return nil, err
	} else if ok {
		return s.ReadFileVersion(ctx, w, id)
	}
	return fs.ReadFile(d, name)
}
