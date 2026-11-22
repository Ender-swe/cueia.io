package workspace

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"

	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/sample"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	DB                                     *pgxpool.Pool
	Provider                               provider.Provider
	Deployment, ProviderName, Region, Mode string
}

func (s *Service) Provision(ctx context.Context, sub string) (string, error) {
	if sub == "" {
		return "", fs.ErrInvalid
	}
	b, err := s.reserve(ctx, sub)
	if err != nil {
		return "", err
	}
	if b.State == "ready" {
		return b.Workspace, nil
	}
	owner := provider.Owner{Deployment: s.Deployment, Workspace: b.Workspace}
	// Serialize provider allocation across API instances without holding the reservation transaction.
	conn, err := s.DB.Acquire(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Release()
	if err := lockProvision(ctx, conn, b.Workspace); err != nil {
		return "", err
	}
	defer unlockProvision(conn, b.Workspace)
	state, err := provisioningState(ctx, conn, b.Workspace)
	if err != nil {
		return "", err
	}
	if state == "ready" {
		return b.Workspace, nil
	}
	m, err := s.Provider.FindOrAllocate(ctx, owner)
	if err != nil {
		return "", err
	}
	d, err := s.Provider.Open(ctx, owner, m)
	if err != nil {
		return "", err
	}
	defer d.Close()
	tx, err := conn.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	if err = lock(ctx, tx, b.Workspace, false); err != nil {
		return "", err
	}
	if err = d.Reconcile(); err != nil {
		return "", err
	}
	if s.Mode == "sample" {
		if err = seed(ctx, tx, b.Workspace, d); err != nil {
			return "", err
		}
	}
	err = markReady(ctx, tx, b.Workspace, m)
	if err != nil {
		return "", err
	}
	return b.Workspace, tx.Commit(ctx)
}

func seed(ctx context.Context, tx pgx.Tx, w string, d provider.Filesystem) error {
	v, err := sample.Load()
	if err != nil {
		return err
	}
	fileID, err := publish(ctx, tx, w, d, "Attachments/"+v.Filename, "application/pdf", sample.PDF)
	if err != nil {
		return err
	}
	return saveSample(ctx, tx, w, fileID, v)
}

func (s *Service) open(ctx context.Context, w string) (provider.Filesystem, error) {
	m, err := s.readyMachine(ctx, w)
	if err != nil {
		return nil, err
	}
	owner := provider.Owner{Deployment: s.Deployment, Workspace: w}
	disk, err := s.Provider.Open(ctx, owner, m)
	var replaced provider.Replacement
	if errors.As(err, &replaced) {
		current := replaced.Current
		if current.DiskID != m.DiskID || current.ID == "" {
			return nil, ErrUnavailable
		}
		disk, err = s.Provider.Open(ctx, owner, current)
		if err == nil {
			if !s.recordReplacement(ctx, w, m, current) {
				disk.Close()
				return nil, ErrUnavailable
			}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("%w: open machine", ErrUnavailable)
	}
	err = pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		if err := lock(ctx, tx, w, false); err != nil {
			return err
		}
		return disk.Reconcile()
	})
	if err != nil {
		disk.Close()
		return nil, err
	}
	return disk, nil
}

func publish(ctx context.Context, tx pgx.Tx, w string, d provider.Filesystem, name, mediaType string, data []byte) (string, error) {
	key, err := d.Publish(name, data)
	if err != nil {
		return "", err
	}
	return recordFileVersion(ctx, tx, w, name, mediaType, key, data)
}

func (s *Service) RegisterFileVersion(ctx context.Context, w, name, mediaType string, validate func([]byte) error) (string, error) {
	if !fs.ValidPath(name) {
		return "", fs.ErrInvalid
	}
	d, err := s.open(ctx, w)
	if err != nil {
		return "", err
	}
	defer d.Close()
	if id, ok, err := s.attachment(ctx, w, name, d); ok || err != nil {
		return id, err
	}
	data, err := fs.ReadFile(d, name)
	if err != nil {
		return "", err
	}
	if validate != nil {
		if err := validate(data); err != nil {
			return "", err
		}
	}
	var id string
	err = pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		if err := lock(ctx, tx, w, true); err != nil {
			return err
		}
		var err error
		id, err = publish(ctx, tx, w, d, name, mediaType, data)
		return err
	})
	return id, err
}

func (s *Service) attachment(ctx context.Context, w, name string, d provider.Filesystem) (string, bool, error) {
	dir, err := attachmentDirectory(d)
	if err != nil {
		return "", false, err
	}
	if path.Dir(name) != dir {
		return "", false, nil
	}
	return s.attachmentID(ctx, w, path.Base(name))
}

func (s *Service) ReadFileVersion(ctx context.Context, w, id string) ([]byte, error) {
	key, err := s.versionKey(ctx, w, id)
	if err != nil {
		return nil, err
	}
	d, err := s.open(ctx, w)
	if err != nil {
		return nil, err
	}
	defer d.Close()
	return d.ReadVersion(key)
}
