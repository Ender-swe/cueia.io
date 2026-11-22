// Package providertest models machine replacement and durable disks in a test directory.
package providertest

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/volume"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider"
)

type Fixture struct {
	Directory string
	// AfterAllocate can fail after persistence to simulate a lost success response.
	AfterAllocate func(provider.Machine) error
	mu            sync.Mutex
}

var _ provider.Provider = (*Fixture)(nil)

func (f *Fixture) ownerDir(owner provider.Owner) (string, error) {
	if owner.Deployment == "" || owner.Workspace == "" {
		return "", errors.New("incomplete workspace identity")
	}
	data, err := json.Marshal(owner)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return filepath.Join(f.Directory, hex.EncodeToString(sum[:])), nil
}

func readBinding(dir string) (m provider.Machine, err error) {
	data, err := os.ReadFile(filepath.Join(dir, "machine.json"))
	if err != nil {
		return m, err
	}
	err = json.Unmarshal(data, &m)
	return m, err
}

func saveBinding(dir string, m provider.Machine) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, "binding-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), filepath.Join(dir, "machine.json"))
}

// Lookup reads an existing binding without allocating a machine.
func (f *Fixture) Lookup(owner provider.Owner) (provider.Machine, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	dir, err := f.ownerDir(owner)
	if err != nil {
		return provider.Machine{}, err
	}
	return readBinding(dir)
}

func (f *Fixture) FindOrAllocate(ctx context.Context, owner provider.Owner) (provider.Machine, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var m provider.Machine
	if err := ctx.Err(); err != nil {
		return m, err
	}
	dir, err := f.ownerDir(owner)
	if err != nil {
		return m, err
	}
	m, err = readBinding(dir)
	if !errors.Is(err, fs.ErrNotExist) {
		return m, err
	}
	if err = os.MkdirAll(filepath.Join(dir, "disk"), 0700); err != nil {
		return m, err
	}
	m = provider.Machine{ID: rand.Text(), DiskID: filepath.Base(dir)}
	if err = saveBinding(dir, m); err != nil {
		return provider.Machine{}, err
	}
	if f.AfterAllocate != nil {
		if err = f.AfterAllocate(m); err != nil {
			return provider.Machine{}, err
		}
	}
	return m, nil
}

func (f *Fixture) checkedDir(ctx context.Context, owner provider.Owner, m provider.Machine) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	dir, err := f.ownerDir(owner)
	if err != nil {
		return "", err
	}
	current, err := readBinding(dir)
	if err != nil {
		return "", err
	}
	if err := provider.CheckBinding(m, current); err != nil {
		return "", err
	}
	return dir, nil
}

func (f *Fixture) Open(ctx context.Context, owner provider.Owner, m provider.Machine) (provider.Filesystem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	dir, err := f.checkedDir(ctx, owner, m)
	if err != nil {
		return nil, err
	}
	return volume.Open(filepath.Join(dir, "disk"))
}

func (f *Fixture) Replace(ctx context.Context, owner provider.Owner, m provider.Machine) (provider.Machine, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	dir, err := f.checkedDir(ctx, owner, m)
	if err != nil {
		return provider.Machine{}, err
	}
	m.ID = rand.Text()
	if err = saveBinding(dir, m); err != nil {
		return provider.Machine{}, err
	}
	return m, nil
}
