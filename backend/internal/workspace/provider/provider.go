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
