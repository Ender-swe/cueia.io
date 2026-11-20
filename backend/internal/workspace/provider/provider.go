// Package provider defines the operations addressed by a durable workspace reservation.
package provider

import (
	"context"
	"io"
	"io/fs"
)

type Owner struct {
	Deployment string
	Workspace  string
}

type Machine struct {
	ID     string
	DiskID string
}

// Replacement reports the current machine only after the provider verifies the
// workspace owner and confirms that its original disk is still attached.
type Replacement struct{ Current Machine }

func (Replacement) Error() string { return "workspace machine replaced" }

func CheckBinding(expected, current Machine) error {
	if expected == current {
		return nil
	}
	if expected.ID != "" && current.ID != "" && expected.DiskID != "" && expected.DiskID == current.DiskID {
		return Replacement{Current: current}
	}
	return fs.ErrPermission
}

type Filesystem interface {
	fs.FS
	io.Closer
	Publish(name string, data []byte) (string, error)
	ReadVersion(key string) ([]byte, error)
	Reconcile() error
	WriteFile(name string, data []byte, perm fs.FileMode) error
}

type Provider interface {
	FindOrAllocate(context.Context, Owner) (Machine, error)
	Open(context.Context, Owner, Machine) (Filesystem, error)
	Replace(context.Context, Owner, Machine) (Machine, error)
}
