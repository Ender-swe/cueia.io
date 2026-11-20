// Package volume separates the browsable tree from service-private document versions on the same disk.
package volume

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Store struct {
	*os.Root
	private *os.Root
	// Tests interrupt a publication at its durability boundaries.
	Interrupt func(string) error
}

func Open(dir string) (*Store, error) {
	for _, sub := range []string{"workspace", "private/staging", "private/objects"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0700); err != nil {
			return nil, err
		}
	}
	root, err := os.OpenRoot(filepath.Join(dir, "workspace"))
	if err != nil {
		return nil, err
	}
	private, err := os.OpenRoot(filepath.Join(dir, "private"))
	if err != nil {
		root.Close()
		return nil, err
	}
	return &Store{Root: root, private: private}, nil
}
func (s *Store) Open(name string) (fs.File, error) { return s.Root.FS().Open(name) }
func (s *Store) Close() error                      { return errors.Join(s.Root.Close(), s.private.Close()) }
func (s *Store) checkpoint(point string) error {
	if s.Interrupt != nil {
		return s.Interrupt(point)
	}
	return nil
}

func Key(path string, data []byte) string {
	return fmt.Sprintf("%x-%x", sha256.Sum256([]byte(path)), sha256.Sum256(data))
}
func validKey(key string) bool {
	if len(key) != 129 || key[64] != '-' {
		return false
	}
	for _, c := range key[:64] + key[65:] {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return false
		}
	}
	return true
}
func (s *Store) ReadVersion(key string) ([]byte, error) {
	if !validKey(key) {
		return nil, fs.ErrInvalid
	}
	b, err := s.private.ReadFile("objects/" + key)
	if err == nil && fmt.Sprintf("%x", sha256.Sum256(b)) != key[65:] {
		return nil, errors.New("stored document checksum mismatch")
	}
	return b, err
}
func (s *Store) Reconcile() error {
	entries, err := fs.ReadDir(s.private.FS(), "staging")
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Type().IsRegular() {
			if err = s.private.Remove("staging/" + e.Name()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) Publish(name string, data []byte) (string, error) {
	if !fs.ValidPath(name) || name == "." || !bytes.HasPrefix(data, []byte("%PDF-")) {
		return "", fs.ErrInvalid
	}
	key := Key(name, data)
	file, err := os.CreateTemp(filepath.Join(s.private.Name(), "staging"), "pdf-")
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err = file.Write(data[:len(data)/2]); err != nil {
		return "", err
	}
	if err = s.checkpoint("staging"); err != nil {
		return "", err
	}
	if _, err = file.Write(data[len(data)/2:]); err != nil {
		return "", err
	}
	if err = file.Sync(); err != nil {
		return "", err
	}
	if err = file.Close(); err != nil {
		return "", err
	}
	source := "staging/" + filepath.Base(file.Name())
	if err = s.private.Link(source, "objects/"+key); err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return "", err
		}
		existing, err := s.ReadVersion(key)
		if err != nil {
			return "", err
		}
		if !bytes.Equal(existing, data) {
			return "", io.ErrUnexpectedEOF
		}
	}
	dir, err := s.private.Open("objects")
	if err != nil {
		return "", err
	}
	err = dir.Sync()
	dir.Close()
	if err != nil {
		return "", err
	}
	if err = s.checkpoint("published"); err != nil {
		return "", err
	}
	return key, s.private.Remove(source)
}
